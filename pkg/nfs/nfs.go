package nfs

import (
	"context"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
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
	if err := create(h.client, conf); err != nil {
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
	if err := delete(h.client, conf); err != nil {
		UpdateStatusPhase(h.client, conf.Name, storagev1.Failed)
		return err
	}
	log.Debugf("[nfs] delete finish")
	return nil
}

func (h *HandlerManager) Update(oldConf, newConf *storagev1.Nfs) error {
	log.Debugf("[nfs] update event, conf: %v", *newConf)
	UpdateStatusPhase(h.client, newConf.Name, storagev1.Updating)
	if err := uMountTmpdir(oldConf.Name); err != nil {
		UpdateStatusPhase(h.client, oldConf.Name, storagev1.Failed)
		return err
	}
	if err := delete(h.client, oldConf); err != nil {
		UpdateStatusPhase(h.client, oldConf.Name, storagev1.Failed)
		return err
	}
	if err := create(h.client, newConf); err != nil {
		UpdateStatusPhase(h.client, newConf.Name, storagev1.Failed)
		return err
	}
	UpdateStatusPhase(h.client, newConf.Name, storagev1.Running)
	log.Debugf("[nfs] update delete finish")
	return nil
}
