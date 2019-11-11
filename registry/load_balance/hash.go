package load_balance

import (
	"fmt"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/emirpasic/gods/utils"
	"github.com/ihaiker/tenured-go-server/registry"

	"errors"
	"hash/crc64"
	"strconv"
)

type HashLoadBalance struct {
	//注册服务的名称
	serverName string

	serverTag string

	//注册中心管理器
	registration registry.ServiceRegistry

	table *crc64.Table

	//虚拟节点数
	virtualNum int

	//保存hash表
	tree *treemap.Map

	//保存注册服务
	serverInstances map[string]*registry.ServerInstance
}

func (this *HashLoadBalance) addSerInstance(instance *registry.ServerInstance) {
	id := instance.Id
	for i := 0; i < this.virtualNum; i++ {
		hashCode := crc64.Checksum([]byte(id+strconv.Itoa(i)), this.table)
		this.tree.Put(hashCode, id)
	}
	this.serverInstances[id] = instance
}

func (this *HashLoadBalance) Start() error {
	ss, err := this.registration.Lookup(this.serverName, []string{this.serverTag})
	if err != nil {
		return err
	}
	for _, serverInstance := range ss {
		this.addSerInstance(serverInstance)
	}

	return this.registration.Subscribe(this.serverName, this.onNotify)
}

func (this *HashLoadBalance) onNotify(serverInstances []*registry.ServerInstance) {
	for _, si := range serverInstances {
		if si.HasTag(this.serverTag) {
			if si.Status == registry.StatusOK {
				this.addSerInstance(si)
			} else if saveIs, has := this.serverInstances[si.Id]; has {
				saveIs.Status = si.Status
			}
		}
	}
}

func (this *HashLoadBalance) Shutdown(interrupt bool) {
	_ = this.registration.Unsubscribe(this.serverName, this.onNotify)
}

func (this *HashLoadBalance) Select(requestCode uint16, obj ...interface{}) ([]*registry.ServerInstance, string, error) {
	if len(obj) == 0 {
		return nil, "", errors.New("not support no param for hash load_balance")
	}

	hashCode := crc64.Checksum([]byte(fmt.Sprintf("%v", obj[0])), this.table)

	_, value := this.tree.Find(func(key interface{}, value interface{}) bool {
		return hashCode <= key.(uint64)
	})
	if value == nil {
		_, value = this.tree.Min()
	}

	serverId := value.(string)
	return []*registry.ServerInstance{this.serverInstances[serverId]}, "", nil
}

func (this *HashLoadBalance) Return(requestCode uint16, key string) {}

func NewHashLoadBalance(serverName string, serverTag string, registration registry.ServiceRegistry, virtualNum int) LoadBalance {
	hlb := &HashLoadBalance{
		serverName: serverName, serverTag: serverTag, registration: registration,

		table: crc64.MakeTable(crc64.ECMA), virtualNum: virtualNum, tree: treemap.NewWith(utils.UInt64Comparator),

		serverInstances: map[string]*registry.ServerInstance{},
	}
	return hlb
}
