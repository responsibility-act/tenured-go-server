package leveldb

import (
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/protocol"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/comparer"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"os"
)

//全局搜索服务
type SearchServer struct {
	dataPath string
	data     *leveldb.DB
}

func (this *SearchServer) Put(key string, value []byte) *protocol.TenuredError {
	if has, err := this.data.Has([]byte(key), readOptions); err != nil {
		return protocol.ErrorDB(err)
	} else if has {
		return api.ErrSearchExists
	}

	if err := this.data.Put([]byte(key), value, writeOptions); err != nil {
		return protocol.ErrorDB(err)
	}
	return nil
}

func (this *SearchServer) Set(key string, body []byte) *protocol.TenuredError {
	if err := this.data.Put([]byte(key), body, writeOptions); err != nil {
		return protocol.ErrorDB(err)
	}
	return nil
}

func (this *SearchServer) Get(key string) ([]byte, *protocol.TenuredError) {
	if value, err := this.data.Get([]byte(key), readOptions); err != nil {
		if err.Error() == levelDBNotFound {
			return nil, api.ErrSearchNotExists
		}
		return nil, protocol.ErrorDB(err)
	} else {
		return value, nil
	}
}

func (this *SearchServer) Remove(key string) *protocol.TenuredError {
	if err := this.data.Delete([]byte(key), writeOptions); err != nil {
		if err.Error() == levelDBNotFound {
			return nil
		}
		return protocol.ErrorDB(err)
	}
	return nil
}

func (this *SearchServer) Start() (err error) {
	logger.Debug("start account store")
	if err = os.MkdirAll(this.dataPath, 0755); err != nil {
		logger.Error("start search store error: ", err)
		return
	}
	if this.data, err = leveldb.OpenFile(this.dataPath, &opt.Options{Comparer: comparer.DefaultComparer}); err != nil {
		logger.Error("start search store error: ", err)
		return err
	}
	return nil
}

func (this *SearchServer) Shutdown(interrupt bool) {
	if err := this.data.Close(); err != nil {
		logger.Error("close search error: ", err)
	}
}

func NewSearchServer(dataPath string) (*SearchServer, error) {
	return &SearchServer{
		dataPath: dataPath + "/store/search",
	}, nil
}
