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
	go StatusControl(h.client, conf.Name)
	var err error
	defer func() {
		if err != nil {
			UpdateStatusPhase(h.client, conf.Name, storagev1.Failed)
		} else {
			UpdateStatusPhase(h.client, conf.Name, storagev1.Running)
		}
	}()
	if err = AddFinalizer(h.client, conf.Name, common.StoragePrestopHookFinalizer); err != nil {
		return err
	}
	if err = create(h.client, conf); err != nil {
		return err
	}
	log.Debugf("[nfs] create finish")
	return nil
}

func (h *HandlerManager) Delete(conf *storagev1.Nfs) error {
	log.Debugf("[nfs] delete event, conf: %v", *conf)
	UpdateStatusPhase(h.client, conf.Name, storagev1.Deleting)
	var err error
	defer func() {
		if err != nil {
			UpdateStatusPhase(h.client, conf.Name, storagev1.Failed)
		}
	}()
	if err = delete(h.client, conf); err != nil {
		return err
	}
	if err = RemoveFinalizer(h.client, conf.Name, common.StoragePrestopHookFinalizer); err != nil {
		return err
	}
	log.Debugf("[nfs] delete finish")
	return nil
}

func (h *HandlerManager) Update(oldConf, newConf *storagev1.Nfs) error {
	log.Debugf("[nfs] update event, conf: %v", *newConf)
	UpdateStatusPhase(h.client, newConf.Name, storagev1.Updating)
	if err := update(h.client, oldConf, newConf); err != nil {
		UpdateStatusPhase(h.client, newConf.Name, storagev1.Failed)
		return err
	}
	UpdateStatusPhase(h.client, newConf.Name, storagev1.Running)
	log.Debugf("[nfs] update delete finish")
	return nil
}
