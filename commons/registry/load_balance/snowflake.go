package load_balance

import (
	"fmt"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/emirpasic/gods/utils"
	"github.com/ihaiker/tenured-go-server/commons/registry"
	"github.com/ihaiker/tenured-go-server/commons/snowflake"
	"hash/crc64"
	"strconv"
)

type SnowflakeExport func(requestCode uint16, parameters ...interface{}) uint64

type element struct {
	Id        string
	StartTime uint64
}

type TimedHashLoadBalance struct {
	//注册服务的名称
	serverName string

	serverTag string

	//注册中心管理器
	registration registry.ServiceRegistry

	snowflakeExport SnowflakeExport

	table *crc64.Table
	//虚拟节点数
	virtualNum int
	//保存hash表
	tree *treemap.Map

	//保存注册服务
	serverInstances map[string]*registry.ServerInstance
}

func (this *TimedHashLoadBalance) addSerInstance(instance *registry.ServerInstance) {
	id := instance.Id
	firstStartTime, _ := strconv.ParseUint(instance.Metadata["FirstStartTime"], 10, 64)

	for i := 0; i < this.virtualNum; i++ {
		hashCode := crc64.Checksum([]byte(id+strconv.Itoa(i)), this.table)
		this.tree.Put(hashCode, &element{Id: id, StartTime: firstStartTime})
	}

	this.serverInstances[id] = instance
}

func (this *TimedHashLoadBalance) Start() error {
	ss, err := this.registration.Lookup(this.serverName, []string{this.serverTag})
	if err != nil {
		return err
	}
	for _, serverInstance := range ss {
		this.addSerInstance(serverInstance)
	}

	return this.registration.Subscribe(this.serverName, this.onNotify)
}

func (this *TimedHashLoadBalance) onNotify(status registry.RegistionStatus, serverInstances []*registry.ServerInstance) {
	for _, si := range serverInstances {
		if status == registry.REGISTER {
			if si.HasTag(this.serverTag) {
				this.addSerInstance(si)
			}
		} else if saveIs, has := this.serverInstances[si.Id]; has {
			saveIs.Status = status.String()
		}
	}
}

func (this *TimedHashLoadBalance) Shutdown(interrupt bool) {
	_ = this.registration.Unsubscribe(this.serverName, this.onNotify)
}

func (this *TimedHashLoadBalance) Select(requestCode uint16, obj ...interface{}) ([]*registry.ServerInstance, string, error) {
	//从请求参数参数中获取分区的snowflake生成的ID
	snowflakeId := this.snowflakeExport(requestCode, obj...)
	//分解此项ID值
	petal := snowflake.Decompose(snowflakeId)

	hashCode := crc64.Checksum([]byte(fmt.Sprintf("%d", snowflakeId)), this.table)

	_, value := this.tree.Find(func(key interface{}, value interface{}) bool {
		return hashCode <= key.(uint64) && (value.(*element).StartTime <= petal.Time)
	})
	if value == nil {
		_, value = this.tree.Find(func(key interface{}, value interface{}) bool {
			return value.(*element).StartTime <= petal.Time
		})
	}
	if value == nil {
		_, value = this.tree.Min()
	}

	serverId := value.(*element).Id
	return []*registry.ServerInstance{this.serverInstances[serverId]}, "", nil
}

func (this *TimedHashLoadBalance) Return(requestCode uint16, key string) {}

func NewTimedHashLoadBalance(serverName string, serverTag string, registration registry.ServiceRegistry, virtualNum int, snowflakeExport SnowflakeExport) LoadBalance {
	hlb := &TimedHashLoadBalance{
		serverName: serverName, serverTag: serverTag, registration: registration,

		snowflakeExport: snowflakeExport, table: crc64.MakeTable(crc64.ECMA),
		virtualNum: virtualNum, tree: treemap.NewWith(utils.UInt64Comparator),

		serverInstances: map[string]*registry.ServerInstance{},
	}
	return hlb
}
