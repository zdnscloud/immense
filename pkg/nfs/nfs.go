package nfs

import (
	"context"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/common"
)

const (
	NfsCSIDpSuffix  = "nfs-client-provisioner"
	NfsDriverSuffix = "nfs.storage.zcloud.cn"
)

var ctx = context.TODO()

type HandlerManager struct {
	client client.Client
}

func New(c client.Client) *HandlerManager {
	return &HandlerManager{
		client: c,
	}
}

func (h *HandlerManager) Create(conf *storagev1.Nfs) error {
	log.Debugf("[nfs] create event, conf: %v", *conf)
	UpdateStatusPhase(h.client, conf.Name, storagev1.Creating)
	if err := deployNfsCSI(h.client, conf); err != nil {
		UpdateStatusPhase(h.client, conf.Name, storagev1.Failed)
		return err
	}
	if err := deployStorageClass(h.client, conf); err != nil {
		UpdateStatusPhase(h.client, conf.Name, storagev1.Failed)
		return err
	}
	if err := AddFinalizer(h.client, conf.Name, common.StoragePrestopHookFinalizer); err != nil {
		UpdateStatusPhase(h.client, conf.Name, storagev1.Failed)
		return err
	}
	UpdateStatusPhase(h.client, conf.Name, storagev1.Running)
	go StatusControl(h.client, conf.Name)
	log.Debugf("[nfs] create finish")
	return nil
}

func (h *HandlerManager) Delete(conf *storagev1.Nfs) error {
	log.Debugf("[nfs] delete event, conf: %v", *conf)
	UpdateStatusPhase(h.client, conf.Name, storagev1.Deleting)
	if err := unDeployNfsCSI(h.client, conf); err != nil {
		UpdateStatusPhase(h.client, conf.Name, storagev1.Failed)
		return err
	}
	if err := unDeployStorageClass(h.client, conf); err != nil {
		UpdateStatusPhase(h.client, conf.Name, storagev1.Failed)
		return err
	}
	if err := RemoveFinalizer(h.client, conf.Name, common.StoragePrestopHookFinalizer); err != nil {
		UpdateStatusPhase(h.client, conf.Name, storagev1.Failed)
		return err
	}
	log.Debugf("[nfs] delete finish")
	return nil
}
