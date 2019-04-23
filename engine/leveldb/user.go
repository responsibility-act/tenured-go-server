package leveldb

import (
	"encoding/json"
	"fmt"
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/api/client"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/protocol"
	"github.com/ihaiker/tenured-go-server/registry"
	"github.com/ihaiker/tenured-go-server/registry/load_balance"
	uuid "github.com/satori/go.uuid"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/comparer"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"os"
	"strconv"
	"strings"
)

func tenantUserKey(accountId, appId uint64, userId string) string {
	return fmt.Sprintf("U:%d:%d:%s", accountId, appId, userId)
}
func cloudKey(accountId, appId uint64, cloudId uint64) []byte {
	return []byte(fmt.Sprintf("C:%d:%d:%d", accountId, appId, cloudId))
}
func tokenKey(accountId, appId uint64, cloudId uint64) []byte {
	return []byte(fmt.Sprintf("T:%d:%d:%d", accountId, appId, cloudId))
}

type UserServer struct {
	storeName string
	dataPath  string
	data      *leveldb.DB

	reg         registry.ServiceRegistry
	loadBalance load_balance.LoadBalance
	search      api.SearchService
	cluster     api.UserService
}

func NewUserServer(serverName, dataPath string) (*UserServer, error) {
	userServer := &UserServer{
		storeName: serverName,
		dataPath:  dataPath + "/store/user",
	}
	return userServer, nil
}

func (this *UserServer) AddUser(user *api.User) *protocol.TenuredError {
	//判断用户是否已经添加
	_, err := this.cluster.GetByTenantUserId(user.AccountId, user.AppId, user.TenantUserId)
	if err == nil || err.Code() != api.ErrUserNotExists.Code() {
		return api.ErrUserExists
	}

	tenantUserKey := tenantUserKey(user.AccountId, user.AppId, user.TenantUserId)
	val := []byte(fmt.Sprintf("%d", user.CloudId))
	if err := this.search.Set(tenantUserKey, val); commons.NotNil(err) {
		if err.Code() == api.ErrSearchExists.Code() {
			return api.ErrUserExists
		}
		return err
	}

	//写入数据
	key := cloudKey(user.AccountId, user.AppId, user.CloudId)
	value, _ := json.Marshal(user)
	if err := this.data.Put(key, value, writeOptions); err != nil {
		return protocol.ErrorDB(err)
	}
	return nil
}

//根据租户给定的用户ID获取用户
func (this *UserServer) GetByTenantUserId(accountId uint64, appId uint64, userId string) (*api.User, *protocol.TenuredError) {
	tenantUserKey := tenantUserKey(accountId, appId, userId)
	if val, err := this.search.Get(tenantUserKey); commons.NotNil(err) {
		if err.Code() == api.ErrSearchNotExists.Code() {
			return nil, api.ErrUserNotExists
		}
		return nil, err
	} else {
		cloudId, _ := strconv.ParseUint(string(val), 10, 64)
		return this.cluster.GetByCloudId(accountId, appId, cloudId)
	}
}

//根据租户给定的用户ID获取用户
func (this *UserServer) GetByCloudId(accountId uint64, appId uint64, cloudId uint64) (*api.User, *protocol.TenuredError) {
	key := cloudKey(accountId, appId, cloudId)
	if val, err := this.data.Get(key, readOptions); err != nil {
		return nil, notFound(nil, api.ErrAccountNotExists)
	} else {
		user := &api.User{}
		_ = json.Unmarshal(val, user)
		return user, nil
	}
}

//更新用户信息，仅允许单个属性更新
func (this *UserServer) ModifyUser(accountId uint64, appId uint64, cloudId uint64, modifyKey string, modifyValue []byte) *protocol.TenuredError {
	user, err := this.GetByCloudId(accountId, appId, cloudId)
	if err != nil {
		return err
	}
	switch modifyKey {
	case "NickName":
		user.NickName = string(modifyKey)
	case "Face":
		user.Face = string(modifyKey)
	default:
		user.Attrs[modifyKey] = string(modifyKey)
	}

	//写入数据
	key := cloudKey(user.AccountId, user.AppId, user.CloudId)
	value, _ := json.Marshal(user)
	if err := this.data.Put(key, value, writeOptions); err != nil {
		return protocol.ErrorDB(err)
	}
	return nil
}

func (this *UserServer) RequestLoginToken(req *api.TokenRequest) (*api.TokenResponse, *protocol.TenuredError) {
	if _, err := this.GetByCloudId(req.AccountId, req.AppId, req.CloudId); err != nil {
		return nil, err
	}
	uuidV4, _ := uuid.NewV4()
	token := &api.TokenResponse{
		Token:  strings.ToUpper(strings.ReplaceAll(uuidV4.String(), "-", "")),
		Linker: "", ExpireTime: req.ExpireTime,
	}
	key := tokenKey(req.AccountId, req.AppId, req.CloudId)
	val, _ := json.Marshal(token)
	if err := this.data.Put(key, val, writeOptions); err != nil {
		return nil, protocol.ErrorDB(err)
	}
	return token, nil
}

func (this *UserServer) GetToken(accountId, appId, cloudId uint64) (*api.TokenResponse, *protocol.TenuredError) {
	key := tokenKey(accountId, appId, cloudId)
	if val, err := this.data.Get(key, readOptions); err != nil {
		return nil, protocol.ErrorDB(err)
	} else {
		token := &api.TokenResponse{}
		_ = json.Unmarshal(val, token)
		return token, nil
	}
}

func (this *UserServer) SetRegistry(serviceRegistry registry.ServiceRegistry) {
	this.reg = serviceRegistry
}

func (this *UserServer) Start() (err error) {
	logger.Debug("start user store.")

	this.loadBalance = NewLoadBalance(this.storeName, this.reg)
	this.search = client.NewSearchServiceClient(this.loadBalance)
	this.cluster = client.NewUserServiceClient(this.loadBalance)

	if err = os.MkdirAll(this.dataPath, 0755); err != nil {
		logger.Error("start account store error: ", err)
		return
	}
	if this.data, err = leveldb.OpenFile(this.dataPath, &opt.Options{Comparer: comparer.DefaultComparer}); err != nil {
		logger.Error("start user store error: ", err)
		return err
	}
	if err = commons.StartIfService(this.search); err != nil {
		return
	}
	if err = commons.StartIfService(this.cluster); err != nil {
		return
	}
	return commons.StartIfService(this.loadBalance)
}

func (this *UserServer) Shutdown(interrupt bool) {
	if err := this.data.Close(); err != nil {
		logger.Error("close user error: ", err)
	}
	commons.ShutdownIfService(this.loadBalance, interrupt)
	commons.ShutdownIfService(this.search, interrupt)
	commons.ShutdownIfService(this.cluster, interrupt)
}
