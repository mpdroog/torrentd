<?php
require_once __DIR__ . '/../shared/curl.php';
use \Curl\Curl;

class GeneralTest extends PHPUnit_Framework_TestCase {
	private static $c;

	public static function setUpBeforeClass() {
		self::$c = new Curl();
		self::$c->setJsonDecoder(function($val) {
			$res = json_decode($val, true);
			return $res;
		});
	}

	public function testNoPath() {
		self::$c->setHeader("Accept", "application/json");
		$res = self::$c->post("http://127.0.0.1:8140", []);
		$this->assertEquals(404, self::$c->httpStatusCode);
	}
}
