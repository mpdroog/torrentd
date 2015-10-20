package torrent

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"encoding/json"
	"net/http"
	"github.com/xsnews/webutils/httpd"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"torrentd/config"
	"os"
	"strings"
)

type Add struct {
	Magnet string // magnet:?xt=urn:btih:1619ecc9373c3639f4ee3e261638f29b33a6cbd6&dn=Ubuntu+14.10+i386+%28Desktop+ISO%29&tr=udp%3A%2F%2Ftracker.openbittorrent.com%3A80&tr=udp%3A%2F%2Fopen.demonii.com%3A1337&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969&tr=udp%3A%2F%2Fexodus.desync.com%3A6969
	Torrent string // d8:announce39:http://torrent.ubuntu.com:6969/announce13:announce-listll39:http://torrent.ubuntu.com:6969/announceel44:http://ipv6.torrent.ubuntu.com:6969...
	Dir string
	User string
}
type List struct {
	User string
}

type ListTorrentRet struct {
	InfoHash string
	BytesCompleted int64
	PieceState []torrent.PieceStateRun
}
type ListRet struct {
	User string
	Torrents []ListTorrentRet
}

var clients map[string]*torrent.Client

func init() {
	clients = make(map[string]*torrent.Client)
}

func Close() error {
	for _, client := range clients {
		client.Close()
	}
	return nil
}

func Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		add(w, r)
		return
	}
	if r.Method == "GET" {
		list(w, r)
		return
	}
	if r.Method == "DELETE" {
		del(w, r)
		return
	}

	w.WriteHeader(404)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("Invalid method."))
}

// Remove torrent from download list
func del(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	if user == "" {
		httpd.Error(w, nil, "Missing user arg.")
		return
	}
	hash := r.URL.Query().Get("hash")
	if hash == "" {
		httpd.Error(w, nil, "Missing hash arg.")
		return
	}

	q, ok := clients[user]
	if !ok {
		httpd.Error(w, nil, "No such user.")
		return
	}

	ok = false
	for _, t := range q.Torrents() {
		if t.InfoHash().HexString() == hash {
			t.Drop()
			ok = true
			break
		}
	}

	if ok {
		w.WriteHeader(200)
	} else {
		w.WriteHeader(404)
	}
}

// List downloads
func list(w http.ResponseWriter, r *http.Request) {
	c := List{}
	c.User = r.URL.Query().Get("user")
	if c.User == "" {
		httpd.Error(w, nil, "Invalid input.")
		return		
	}
	q, ok := clients[c.User]
	if !ok {
		httpd.Error(w, nil, "No such user.")
		return
	}

	res := ListRet{User: c.User}
	for _, t := range q.Torrents() {
		res.Torrents = append(res.Torrents, ListTorrentRet{
			InfoHash: t.InfoHash().HexString(),
			BytesCompleted: t.BytesCompleted(),
			PieceState: t.PieceStateRuns(),
		})
	}

	j, e := json.Marshal(res)
	if e != nil {
		httpd.Error(w, nil, "Failed json encoding.")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, e := w.Write(j); e != nil {
		httpd.Error(w, e, "Flush failed")
		return
	}
}

// Add Torrent/Magnet
func add(w http.ResponseWriter, r *http.Request) {
	c := Add{}
	if e := json.NewDecoder(r.Body).Decode(&c); e != nil {
		httpd.Error(w, e, "Invalid input.")
		return
	}
	if c.User == "" {
		httpd.Error(w, nil, "No user.")
		return
	}
	if c.Dir == "" {
		httpd.Error(w, nil, "No dir.")
		return
	}
	if strings.Contains(c.Dir, "..") {
		httpd.Error(w, nil, "Dir hack.")
		return
	}

	if _, ok := clients[c.User]; !ok {
		dlDir := fmt.Sprintf(config.C.Basedir + c.Dir)
		if _, e := os.Stat(config.C.Basedir); os.IsNotExist(e) {
			if e := os.Mkdir(dlDir, os.ModeDir); e != nil {
				config.L.Printf("CRIT: %s", e.Error())
				httpd.Error(w, e, "Permission error.")
				return
			}
		}

		// https://github.com/anacrolix/torrent/blob/master/config.go#L9
		cl, e := torrent.NewClient(&torrent.Config{
			DataDir: config.C.Basedir + c.Dir,
			// IPBlocklist => http://john.bitsurge.net/public/biglist.p2p.gz
		})
		if e != nil {
			httpd.Error(w, e, "Client init failed")
			return
		}
		clients[c.User] = cl
	}

	client := clients[c.User]
	var (
		t torrent.Torrent
		e error
	)
	if len(c.Magnet) > 0 {
		t, e = client.AddMagnet(c.Magnet)
		if e != nil {
			httpd.Error(w, e, "Magnet add failed")
			return
		}
	} else if len(c.Torrent) > 0 {
		// strip base64
		b, e := base64.StdEncoding.DecodeString(c.Torrent)
		if e != nil {
			httpd.Error(w, e, "Failed base64 decode torrent input")
			return
		}
		m, e := metainfo.Load(bytes.NewReader(b))
		if e != nil {
			httpd.Error(w, e, "Failed base64 decode torrent input")
			return
		}
		t, e = client.AddTorrent(m)
		if e != nil {
			httpd.Error(w, e, "Failed adding torrent")
			return
		}
	} else {
		httpd.Error(w, nil, "No magnet nor torrent.")
		return
	}

	// queue
	go func() {
		<-t.GotInfo()
		t.DownloadAll()
	}()

	// (cl *Client) AddTorrentSpec(spec *TorrentSpec) (T Torrent, new bool, err error) {
	// TorrentSpecFromMagnetURI(uri string) (spec *TorrentSpec, err error) {
	// TorrentSpecFromMetaInfo(mi *metainfo.MetaInfo) (spec *TorrentSpec) {
	// (me *Client) AddMagnet(uri string) (T Torrent, err error) {
	// (me *Client) AddTorrent(mi *metainfo.MetaInfo) (T Torrent, err error) {

	msg := fmt.Sprintf(`{"status": true, "hash": "%s", "text": "Queued."}`, t.InfoHash().HexString())
	w.Header().Set("Content-Type", "application/json")
	if _, e := w.Write([]byte(msg)); e != nil {
		httpd.Error(w, e, "Flush failed")
		return
	}
}
