<?php
require_once __DIR__ . '/../shared/curl.php';
use \Curl\Curl;

class ListTorrentTest extends PHPUnit_Framework_TestCase {
	private static $c;

	public static function setUpBeforeClass() {
		self::$c = new Curl();
		self::$c->setJsonDecoder(function($val) {
			$res = json_decode($val, true);
			return $res;
		});
	}

	public function testMissingUser() {
		self::$c->setHeader("Accept", "application/json");
		$res = self::$c->get("http://127.0.0.1:8140/torrent", []);
		$this->assertEquals(500, self::$c->httpStatusCode);
		$this->assertEquals(["status" => false, "text" => "Invalid input."], $res);		
	}

	public function testEmptyUser() {
		self::$c->setHeader("Accept", "application/json");
		$res = self::$c->get("http://127.0.0.1:8140/torrent?user=notexist");
		$this->assertEquals(500, self::$c->httpStatusCode);
		$this->assertEquals(["status" => false, "text" => "No such user."], $res);
	}

	public function testAddMagnet() {
		self::$c->setHeader("Accept", "application/json");
		$res = self::$c->post("http://127.0.0.1:8140/torrent", json_encode([
			"magnet" => "magnet:?xt=urn:btih:1619ecc9373c3639f4ee3e261638f29b33a6cbd6&dn=Ubuntu+14.10+i386+%28Desktop+ISO%29&tr=udp%3A%2F%2Ftracker.openbittorrent.com%3A80&tr=udp%3A%2F%2Fopen.demonii.com%3A1337&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969&tr=udp%3A%2F%2Fexodus.desync.com%3A6969",
			"user" => "herpie",
			"dir" => "herpie"
		]));
		$this->assertEquals(200, self::$c->httpStatusCode);
		$this->assertEquals([
			"status" => true,
			"hash" => "1619ecc9373c3639f4ee3e261638f29b33a6cbd6",
			"text" => "Queued."
		], $res);
	}

	public function testList() {
		self::$c->setHeader("Accept", "application/json");
		$res = self::$c->get("http://127.0.0.1:8140/torrent?user=herpie");
		$this->assertEquals(200, self::$c->httpStatusCode);
		$this->assertEquals([
			"User" => "herpie",
			"Torrents" => [[
				"InfoHash" => "1619ecc9373c3639f4ee3e261638f29b33a6cbd6",
				"BytesCompleted" => 0,
				"PieceState" => null
			]]
		], $res);
	}

	public function testAddTorrent() {
		self::$c->setHeader("Accept", "application/json");
		$res = self::$c->post("http://127.0.0.1:8140/torrent", json_encode([
			"torrent" => base64_encode(file_get_contents("./ubuntu-14.10-server-amd64.iso.torrent")),
			"user" => "herpie",
			"dir" => "herpie"
		]));
		$this->assertEquals(200, self::$c->httpStatusCode);
		$this->assertEquals([
			"status" => true,
			"hash" => "ec5ce22050e0e3d7e7a279c241901b5bc1f36fe7",
			"text" => "Queued."
		], $res);
	}

	public function testList2() {
		$sortedRes = [];
		self::$c->setHeader("Accept", "application/json");
		$res = self::$c->get("http://127.0.0.1:8140/torrent?user=herpie");
		$this->assertEquals(200, self::$c->httpStatusCode);

		$sortedRes = $res;
		$sortedRes["Torrents"] = [];
		foreach ($res["Torrents"] as $torrent) {
			if ($torrent["InfoHash"] === "1619ecc9373c3639f4ee3e261638f29b33a6cbd6") {
				$sortedRes["Torrents"][0] = $torrent;
			} else {
				$sortedRes["Torrents"][1] = $torrent;				
			}
		}

		$this->assertEquals([
			"User" => "herpie",
			"Torrents" => [
			[
				"InfoHash" => "1619ecc9373c3639f4ee3e261638f29b33a6cbd6",
				"BytesCompleted" => 0,
				"PieceState" => null
			], [
				"InfoHash" => "ec5ce22050e0e3d7e7a279c241901b5bc1f36fe7",
				"BytesCompleted" => 0,
				"PieceState" => [
					["Priority" => 2, "Complete" => false, "Checking" => false, "Partial" => false, "Length" => 1],
					["Priority" => 1, "Complete" => false, "Checking" => false, "Partial" => false, "Length" => 1162],
					["Priority" => 2, "Complete" => false, "Checking" => false, "Partial" => false, "Length" => 1],
				]
			]]
		], $sortedRes);
	}

	public function testDelete() {
		{
			self::$c->setHeader("Accept", "application/json");
			$res = self::$c->delete("http://127.0.0.1:8140/torrent?user=herpie&hash=1619ecc9373c3639f4ee3e261638f29b33a6cbd6");
			$this->assertEquals(200, self::$c->httpStatusCode);
		}

		{
			self::$c->setHeader("Accept", "application/json");
			$res = self::$c->get("http://127.0.0.1:8140/torrent?user=herpie");
			$this->assertEquals(200, self::$c->httpStatusCode);
			$this->assertEquals([
				"User" => "herpie",
				"Torrents" => [
				[
					"InfoHash" => "ec5ce22050e0e3d7e7a279c241901b5bc1f36fe7",
					"BytesCompleted" => 0,
					"PieceState" => [
						["Priority" => 2, "Complete" => false, "Checking" => false, "Partial" => false, "Length" => 1],
						["Priority" => 1, "Complete" => false, "Checking" => false, "Partial" => false, "Length" => 1162],
						["Priority" => 2, "Complete" => false, "Checking" => false, "Partial" => false, "Length" => 1],
					]
				]]
			], $res);
		}
	}
}
