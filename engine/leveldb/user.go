package leveldb

import (
	"encoding/json"
	"fmt"
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/executors"
	"github.com/ihaiker/tenured-go-server/commons/protocol"
	"github.com/ihaiker/tenured-go-server/commons/registry"
	"github.com/ihaiker/tenured-go-server/commons/registry/load_balance"
	"github.com/ihaiker/tenured-go-server/commons/remoting"
	uuid "github.com/satori/go.uuid"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/comparer"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"os"
	"strconv"
	"strings"
	"time"
)

func tenantUserKey(accountId, appId uint64, userId string) []byte {
	return []byte(fmt.Sprintf("U:%d:%d:%s", accountId, appId, userId))
}
func clusterUserKey(accountId, appId uint64, clusterId uint64) []byte {
	return []byte(fmt.Sprintf("C:%d:%d:%d", accountId, appId, clusterId))
}
func tokenKey(accountId, appId uint64, clusterId uint64) []byte {
	return []byte(fmt.Sprintf("T:%d:%d:%d", accountId, appId, clusterId))
}

type UserServer struct {
	dataPath string
	data     *leveldb.DB

	server      *protocol.TenuredServer
	client      *protocol.TenuredClient
	manager     executors.ExecutorManager
	loadBalance load_balance.LoadBalance
}

func NewUserServer(dataPath string, serverName string, reg registry.ServiceRegistry, server *protocol.TenuredServer, manager executors.ExecutorManager) (*UserServer, error) {
	userServer := &UserServer{
		dataPath: dataPath + "/store/user",
		server:   server,
		manager:  manager,
	}
	userServer.loadBalance = UserLoadBalance(serverName, api.StoreUser, reg)

	config := remoting.DefaultConfig()
	config.SendLimit = 100
	if client, err := protocol.NewTenuredClient(config); err != nil {
		return nil, err
	} else {
		userServer.client = client
	}
	return userServer, nil
}

func (this *UserServer) AddUser(user *api.User) *protocol.TenuredError {
	if _, err := this.GetByTenantUserId(user.AccountId, user.AppId, user.TenantUserId); err != api.ErrUserNotExists {
		return api.ErrUserExists
	}
	key := tenantUserKey(user.AccountId, user.AppId, user.TenantUserId)
	value := fmt.Sprintf("%d", user.ClusterId)
	if err := this.data.Put(key, []byte(value), writeOptions); err != nil {
		return protocol.ErrorDB(err)
	}
	return this.syncSetUser(user)
}

//根据租户给定的用户ID获取用户
func (this *UserServer) GetByTenantUserId(accountId uint64, appId uint64, userId string) (*api.User, *protocol.TenuredError) {
	key := tenantUserKey(accountId, appId, userId)
	if val, err := this.data.Get(key, readOptions); err != nil {
		if err.Error() == levelDBNotFound {
			return nil, api.ErrUserNotExists
		} else {
			return nil, protocol.ErrorDB(err)
		}
	} else {
		clusterId, _ := strconv.ParseUint(string(val), 10, 64)
		return this.syncGetUser(accountId, appId, clusterId)
	}
}

//根据租户给定的用户ID获取用户
func (this *UserServer) GetByClusterId(accountId uint64, appId uint64, clusterId uint64) (*api.User, *protocol.TenuredError) {
	key := clusterUserKey(accountId, appId, clusterId)
	if val, err := this.data.Get(key, readOptions); err != nil {
		if err.Error() == levelDBNotFound {
			return nil, api.ErrUserNotExists
		} else {
			return nil, protocol.ErrorDB(err)
		}
	} else {
		user := &api.User{}
		if err := json.Unmarshal(val, user); err != nil {
			return nil, protocol.ErrorDB(err)
		}
		return user, nil
	}
}

//更新用户信息，仅允许单个属性更新
func (this *UserServer) ModifyUser(accountId uint64, appId uint64, clusterId uint64, modifyKey string, modifyValue []byte) *protocol.TenuredError {
	user, err := this.GetByClusterId(accountId, appId, clusterId)
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
	return nil
}

func (this *UserServer) RequestLoginToken(requestToken *api.TokenRequest) (*api.TokenResponse, *protocol.TenuredError) {
	_, err := this.GetByClusterId(requestToken.AccountId, requestToken.AppId, requestToken.ClusterId)
	if err != nil {
		return nil, err
	}
	uuidV4, _ := uuid.NewV4()

	token := &api.TokenResponse{
		Token:  strings.ToUpper(strings.ReplaceAll(uuidV4.String(), "-", "")),
		Linker: "", ExpireTime: requestToken.ExpireTime,
	}

	key := tokenKey(requestToken.AccountId, requestToken.AppId, requestToken.ClusterId)
	val, _ := json.Marshal(token)
	if err := this.data.Put(key, val, writeOptions); err != nil {
		return nil, protocol.ErrorDB(err)
	}
	return token, nil
}

func (this *UserServer) GetToken(accountId, appId, clusterId uint64) (*api.TokenResponse, *protocol.TenuredError) {
	key := tokenKey(accountId, appId, clusterId)
	if val, err := this.data.Get(key, readOptions); err != nil {
		return nil, protocol.ErrorDB(err)
	} else {
		token := &api.TokenResponse{}
		_ = json.Unmarshal(val, token)
		return token, nil
	}
}

func (this *UserServer) selectAddress(accountId, appId, clusterId uint64) (string, *protocol.TenuredError) {
	serverInstances, retKey, err := this.loadBalance.Select(api.UserServiceGetByClusterId, accountId, appId, clusterId)
	if err != nil || len(serverInstances) == 0 || registry.AllNotOK(serverInstances...) {
		return "", protocol.ErrorRouter()
	}
	defer this.loadBalance.Return(api.UserServiceGetByClusterId, retKey)
	return serverInstances[0].Address, nil
}

func (this *UserServer) syncSetUser(user *api.User) *protocol.TenuredError {
	address, err := this.selectAddress(user.AccountId, user.AppId, user.ClusterId)
	if !commons.IsNil(err) {
		return protocol.ConvertError(err)
	}
	value, _ := json.Marshal(user)
	//就是本地服务
	if this.server.Address == address {
		key := clusterUserKey(user.AccountId, user.AppId, user.ClusterId)
		if err := this.data.Put(key, value, writeOptions); err != nil {
			return protocol.ErrorDB(err)
		}
		return nil
	}

	request := protocol.NewRequest(api.UserServiceRange.Max)
	request.Body = value
	if response, err := this.client.Invoke(address, request, time.Millisecond*3000); err != nil {
		return protocol.ConvertError(err)
	} else if !response.IsSuccess() {
		return response.GetError()
	}
	return nil
}

func (this *UserServer) syncGetUser(accountId, appId, clusterId uint64) (*api.User, *protocol.TenuredError) {
	address, err := this.selectAddress(accountId, appId, clusterId)
	if !commons.IsNil(err) {
		return nil, protocol.ConvertError(err)
	}
	if this.server.Address == address {
		return this.GetByClusterId(accountId, appId, clusterId)
	}

	request := protocol.NewRequest(api.UserServiceGetByClusterId)
	_ = request.SetHeader(&struct {
		AccountId uint64 `json:"accountId"`
		AppId     uint64 `json:"appId"`
		ClusterId uint64 `json:"clusterId"`
	}{
		AccountId: accountId,
		AppId:     appId,
		ClusterId: clusterId,
	})
	user := &api.User{}
	if response, err := this.client.Invoke(address, request, time.Millisecond*3000); !commons.IsNil(err) {
		return nil, protocol.ConvertError(err)
	} else if !response.IsSuccess() {
		return nil, response.GetError()
	} else {
		_ = response.GetHeader(user)
		return user, nil
	}
}

func (this *UserServer) handlerSyncUser() error {
	this.server.RegisterCommandProcesser(api.UserServiceRange.Max, func(channel remoting.RemotingChannel, request *protocol.TenuredCommand) {
		user := &api.User{}
		_ = json.Unmarshal(request.Body, user)
		key := clusterUserKey(user.AccountId, user.AppId, user.ClusterId)

		response := protocol.NewACK(request.ID())
		if err := this.data.Put(key, request.Body, writeOptions); err != nil {
			response.RemotingError(protocol.ErrorDB(err))
		}
		if err := channel.Write(response, time.Second*3); err != nil {
			logger.Warn("sync set user response error", err)
		}
	}, this.manager.Get("User.SyncSetUser"))
	return nil
}

func (this *UserServer) Start() (err error) {
	logger.Debug("start user store.")
	if err = os.MkdirAll(this.dataPath, 0755); err != nil {
		logger.Error("start account store error: ", err)
		return
	}
	if this.data, err = leveldb.OpenFile(this.dataPath, &opt.Options{Comparer: comparer.DefaultComparer}); err != nil {
		logger.Error("start user store error: ", err)
		return err
	}
	return this.handlerSyncUser()
}

func (this *UserServer) Shutdown(interrupt bool) {
	if err := this.data.Close(); err != nil {
		logger.Error("close user error: ", err)
	}
}
