package main

import (
	"fmt"
	"net"
	"net/http"
	"torrentd/config"
	"torrentd/torrent"
	"github.com/xsnews/webutils/httpd"
	"github.com/xsnews/webutils/middleware"
	"github.com/xsnews/webutils/muxdoc"
	"github.com/xsnews/webutils/ratelimit"	
)

var (
	mux muxdoc.MuxDoc
	ln net.Listener
)

func Control() {
	mux.Title = "torrentD API"
	mux.Desc = "Administrative API to control TorrentD"
	mux.Add("/", doc, "This documentation")
	mux.Add("/shutdown", shutdown, "Finish jobs and close application")
	mux.Add("/verbose", verbose, "Toggle verbosity-mode")
	mux.Add("/torrent", torrent.Handle, "Interact on down/uploads.")

	middleware.Add(ratelimit.Use(50, 50))
	http.Handle("/", middleware.Use(mux.Mux))

	var e error
	server := &http.Server{Addr: config.C.Listen, Handler: nil}
	ln, e = net.Listen("tcp", server.Addr)
	if e != nil {
		panic(e)
	}
	if config.Verbose {
		config.L.Printf("torrentd listening on %s", config.C.Listen)
	}
	if e := server.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)}); e != nil {
		if !config.Stopping {
			panic(e)
		}
	}
}

// Return API Documentation (paths)
func doc(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(mux.String()))
}

// Finish pending jobs and close application
func shutdown(w http.ResponseWriter, r *http.Request) {
	if config.Stopping {
		pending := 0
		if _, e := w.Write([]byte(fmt.Sprintf(`{success: true, msg: "Already stopping.", pending: %d}`, pending))); e != nil {
			httpd.Error(w, e, "Flush failed")
			return
		}
	}
	config.L.Printf("Disconnecting")
	config.Stopping = true

	if e := ln.Close(); e != nil {
		httpd.Error(w, e, `{success: false, msg: "Error stopping HTTP-listener"}`)
	}
	if e := torrent.Close(); e != nil {
		httpd.Error(w, e, `{success: false, msg: "Error stopping Torrent-IO"}`)		
	}

	if _, e := w.Write([]byte(`{success: true, msg: "Stopped listening, waiting for empty queue."}`)); e != nil {
		httpd.Error(w, e, "Flush failed")
		return
	}
}

func verbose(w http.ResponseWriter, r *http.Request) {
	msg := `{success: true, msg: "Set verbosity to `
	if config.Verbose {
		config.Verbose = false
		msg += "OFF"
	} else {
		config.Verbose = true
		msg += "ON"
	}
	msg += `"}`

	if _, e := w.Write([]byte(msg)); e != nil {
		httpd.Error(w, e, "Flush failed")
		return
	}
}
