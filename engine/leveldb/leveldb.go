package leveldb

import "github.com/syndtr/goleveldb/leveldb/opt"

var readOptions = &opt.ReadOptions{}
var writeOptions = &opt.WriteOptions{Sync: true}

const levelDBNotFound = "leveldb: not found"
