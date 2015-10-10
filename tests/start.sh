#!/bin/bash
# Blackbox testing
#
set +x
TORRENTD=0
function cleanup {
	# Stop PHP webserver
	kill -s TERM $TORRENTD
}
trap cleanup EXIT

cd ..
./torrentd -v &
TORRENTD=$!
cd -

phpunit tests
