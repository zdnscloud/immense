package iscsi

import (
	"context"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/common"
)

const (
	StorageTypeSuffix           = "iscsi"
	IscsiInstanceSecretSuffix   = "iscsi-secret"
	IscsiDriverSuffix           = "iscsi.storage.zcloud.cn"
	IscsiCSIDsSuffix            = "iscsi-plugin"
	IscsiCSIDpSuffix            = "iscsi-csi"
	IscsiLvmdDsSuffix           = "iscsi-lvmd"
	IscsiStopJobSuffix          = "iscsi-stop-job"
	VolumeGroupSuffix           = "iscsi-group"
	IscsiInstanceLabelKeyPrefix = "storage.zcloud.cn/iscsi-instance"
	IscsiInstanceLabelValue     = "true"
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

func (h *HandlerManager) Create(conf *storagev1.Iscsi) error {
	log.Debugf("[iscsi] create event, conf: %v", *conf)
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
	log.Debugf("[iscsi] create finish")
	return nil
}

func (h *HandlerManager) Update(oldConf, newConf *storagev1.Iscsi) error {
	log.Debugf("[iscsi] update event, old: %v,new: %v", *oldConf, *newConf)
	UpdateStatusPhase(h.client, newConf.Name, storagev1.Updating)
	if err := update(h.client, oldConf, newConf); err != nil {
		UpdateStatusPhase(h.client, newConf.Name, storagev1.Failed)
		return err
	}
	UpdateStatusPhase(h.client, newConf.Name, storagev1.Running)
	log.Debugf("[iscsi] update finish")
	return nil
}

func (h *HandlerManager) Delete(conf *storagev1.Iscsi) error {
	log.Debugf("[iscsi] delete event, conf: %v", *conf)
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
	log.Debugf("[iscsi] delete finish")
	return nil
}

func (h *HandlerManager) Redeploy(name string) error {
	log.Debugf("[iscsi] redeploy %s", name)
	UpdateStatusPhase(h.client, name, storagev1.Updating)
	var err error
	defer func() {
		if err != nil {
			UpdateStatusPhase(h.client, name, storagev1.Failed)
		} else {
			UpdateStatusPhase(h.client, name, storagev1.Running)
		}
	}()
	var iscsi *storagev1.Iscsi
	iscsi, err = getIscsi(h.client, name)
	if err != nil {
		UpdateStatusPhase(h.client, name, storagev1.Failed)
		return err
	}
	if err = redeploy(h.client, iscsi, iscsi); err != nil {
		UpdateStatusPhase(h.client, name, storagev1.Failed)
		return err
	}
	log.Debugf("[iscsi] redeploy finish")
	return nil
}
