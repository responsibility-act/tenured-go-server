package leveldb

import (
	"github.com/ihaiker/tenured-go-server/protocol"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

var readOptions = &opt.ReadOptions{}
var writeOptions = &opt.WriteOptions{Sync: true}

const levelDBNotFound = "leveldb: not found"

func notFound(err error, perr *protocol.TenuredError) *protocol.TenuredError {
	if err.Error() == levelDBNotFound {
		return perr
	} else {
		return protocol.ErrorDB(err)
	}
}
