package eventhandler

import (
	"errors"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/ceph"
	"github.com/zdnscloud/immense/pkg/common"
	"github.com/zdnscloud/immense/pkg/lvm"
	"reflect"
)

type Handler interface {
	GetType() string
	Create(cluster common.Storage) error
	Delete(cluster common.Storage) error
	Update(oldc, newc common.Storage) error
}

type HandlerManager struct {
	handlers []Handler
	client   client.Client
}

func New(cli client.Client) *HandlerManager {
	return &HandlerManager{
		handlers: []Handler{
			lvm.New(cli),
			ceph.New(cli),
		},
		client: cli,
	}
}

func (h *HandlerManager) Create(cluster *storagev1.Cluster) error {
	for _, s := range h.handlers {
		if s.GetType() == cluster.Spec.StorageType {
			log.Debugf("[%s] create event", cluster.Spec.StorageType)
			newcluster, err := common.AssembleCreateConfig(h.client, cluster)
			if err != nil {
				return err
			}
			logCluster(newcluster, "create")
			return s.Create(newcluster)
		}
	}
	return nil
}

func (h *HandlerManager) Delete(cluster *storagev1.Cluster) error {
	for _, s := range h.handlers {
		if s.GetType() == cluster.Spec.StorageType {
			log.Debugf("[%s] delete event", cluster.Spec.StorageType)
			newcluster, err := common.AssembleDeleteConfig(h.client, cluster)
			if err != nil {
				return err
			}
			logCluster(newcluster, "delete")
			return s.Delete(newcluster)
		}
	}
	return nil
}

func (h *HandlerManager) Update(oldc *storagev1.Cluster, newc *storagev1.Cluster) error {
	if oldc.Name != newc.Name || oldc.Spec.StorageType != newc.Spec.StorageType {
		return errors.New("Invalid config!")
	}
	if reflect.DeepEqual(oldc.Spec.Hosts, newc.Spec.Hosts) {
		return nil
	}
	for _, s := range h.handlers {
		if s.GetType() == oldc.Spec.StorageType {
			log.Debugf("[%s] update event", newc.Spec.StorageType)
			dels, adds, err := common.AssembleUpdateConfig(h.client, oldc, newc)
			if err != nil {
				return err
			}
			if len(dels.Spec.Hosts) > 0 {
				logCluster(dels, "delete")
			}
			if len(adds.Spec.Hosts) > 0 {
				logCluster(adds, "create")
			}
			return s.Update(dels, adds)
		}
	}
	return nil
}

func logCluster(cluster common.Storage, action string) {
	log.Debugf("name:%s, type:%s, action:%s", cluster.Name, cluster.Spec.StorageType, action)
	for _, host := range cluster.Spec.Hosts {
		log.Debugf("node:%s, devs:%s", host.NodeName, host.BlockDevices)
	}
}
