package services

import (
	"fmt"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/registry"
	"github.com/kataras/iris/core/errors"
	"github.com/satori/go.uuid"
	"strings"
)

//生成一个集群唯一ID，此ID全局唯一且为顺序递增。
// 组成方式: version(0-F) + zoneId (0000-FFFF) + UUID
type ClusterID struct {
	//注册中心实现
	reg registry.ServiceRegistry
	//工作目录
	workDir string

	clusterIdFile *commons.File
}

func (this *ClusterID) Id(serverName string) (string, error) {
	this.clusterIdFile = commons.NewFile(this.workDir + fmt.Sprintf("/%s.cid", serverName))
	if this.clusterIdFile.Exist() {
		return this.clusterIdFile.ToString()
	}

	if ss, err := this.reg.Lookup(serverName, nil); err != nil {
		return "", err
	} else {
		clusterID := len(ss)
		uuidObj, _ := uuid.NewV4()
		uuidString := strings.ReplaceAll(uuidObj.String(), "-", "")
		id := fmt.Sprintf("1-%04X-%s", clusterID, uuidString)
		return id, nil
	}
}

func (this *ClusterID) CheckAndWrite(serverName string, clusterId string) error {
	if ss, err := this.reg.Lookup(serverName, nil); err != nil {
		return err
	} else {
		clusterZone := map[string]interface{}{}
		for _, v := range ss {
			clusterId := v.Id
			zoneId := clusterId[2:6]
			if _, has := clusterZone[zoneId]; has {
				return errors.New("double or more zoneId: " + zoneId)
			}
		}
	}

	if writer, err := this.clusterIdFile.GetWriter(false); err != nil {
		return err
	} else {
		defer func() { _ = writer.Close() }()
		_, err := writer.WriteString(clusterId)
		return err
	}
}

func NewClusterID(workDir string, reg registry.ServiceRegistry) *ClusterID {
	return &ClusterID{
		workDir: workDir,
		reg:     reg,
	}
}
