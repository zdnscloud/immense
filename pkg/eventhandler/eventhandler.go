package eventhandler

import (
	"errors"
	"reflect"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/ceph"
	"github.com/zdnscloud/immense/pkg/common"
	"github.com/zdnscloud/immense/pkg/lvm"
)

type Handler interface {
	GetType() string
	Create(cluster storagev1.Cluster) error
	Delete(cluster storagev1.Cluster) error
	Update(oldc, newc storagev1.Cluster) error
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
			log.Debugf("create event for storage type %s", cluster.Spec.StorageType)
			newcluster, err := common.AssembleCreateConfig(h.client, cluster)
			if err != nil {
				common.UpdateStatusPhase(h.client, cluster.Name, storagev1.Failed)
				return err
			}
			logCluster(*newcluster, "create")
			if err := s.Create(*newcluster); err != nil {
				return err
			}
			log.Debugf("create storage type %s finish", cluster.Spec.StorageType)
		}
	}
	return nil
}

func (h *HandlerManager) Delete(cluster *storagev1.Cluster) error {
	for _, s := range h.handlers {
		if s.GetType() == cluster.Spec.StorageType {
			log.Debugf("delete event for storage type %s", cluster.Spec.StorageType)
			logCluster(*cluster, "delete")
			if err := s.Delete(*cluster); err != nil {
				return err
			}
			log.Debugf("delete storage type %s finish", cluster.Spec.StorageType)
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
			log.Debugf("update event for storage type %s", oldc.Spec.StorageType)
			dels, adds, err := common.AssembleUpdateConfig(h.client, oldc, newc)
			if err != nil {
				common.UpdateStatusPhase(h.client, oldc.Name, storagev1.Failed)
				return err
			}
			if len(dels.Spec.Hosts) > 0 {
				logCluster(*dels, "delete")
			}
			if len(adds.Spec.Hosts) > 0 {
				logCluster(*adds, "create")
			}
			if err := s.Update(*dels, *adds); err != nil {
				return err
			}
			log.Debugf("update storage type %s finish", oldc.Spec.StorageType)
		}
	}
	return nil
}

func logCluster(cluster storagev1.Cluster, action string) {
	log.Debugf("name:%s, type:%s, action:%s", cluster.Name, cluster.Spec.StorageType, action)
	for _, host := range cluster.Status.Config {
		devs := make([]string, 0)
		for _, dev := range host.BlockDevices {
			devs = append(devs, dev)
		}
		log.Debugf("node:%s, devs:%s", host.NodeName, devs)
	}
}
