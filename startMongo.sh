#! /bin/sh

mongod --config /usr/local/etc/mongod.conf --wiredTigerCacheSizeGB 0.4

# for server: mongod --config /etc/mongod.conf --wiredTigerCacheSizeGB 0.4 &