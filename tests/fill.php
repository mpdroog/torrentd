<?php
require_once __DIR__ . '/shared/curl.php';
use \Curl\Curl;

$c = new Curl();
$c->setJsonDecoder(function($val) {
	$res = json_decode($val, true);
	return $res;
});

$c->setHeader("Accept", "application/json");
$res = $c->post("http://127.0.0.1:8140/torrent", json_encode([
	"torrent" => base64_encode(file_get_contents("./ubuntu-14.10-server-amd64.iso.torrent")),
	"user" => "herpie",
	"dir" => "herpie"
]));
var_dump($res);
