package lvm

import (
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/common"
)

type Lvm struct {
	cli client.Client
}

const (
	StorageType           = "lvm"
	VolumeGroup           = "k8s"
	LvmdDsName            = "lvmd"
	CSIPluginDsName       = "csi-lvmplugin"
	CSIProvisionerStsName = "csi-lvmplugin-provisioner"
	LvmDriverSuffix       = "lvm.storage.zcloud.cn"
)

func New(c client.Client) *Lvm {
	return &Lvm{
		cli: c,
	}
}

func (s *Lvm) GetType() string {
	return StorageType
}

func (s *Lvm) Create(cluster storagev1.Cluster) error {
	common.UpdateClusterStatusPhase(s.cli, cluster.Name, storagev1.Creating)
	go StatusControl(s.cli, cluster.Name)
	var err error
	defer func() {
		if err != nil {
			common.UpdateClusterStatusPhase(s.cli, cluster.Name, storagev1.Failed)
		} else {
			common.UpdateClusterStatusPhase(s.cli, cluster.Name, storagev1.Running)
		}
	}()
	if err = common.AddFinalizerForStorage(s.cli, cluster.Name, common.StoragePrestopHookFinalizer); err != nil {
		return err
	}
	if err = common.CreateNodeAnnotationsAndLabels(s.cli, common.StorageHostLabels, s.GetType(), cluster.Spec.Hosts); err != nil {
		return err
	}
	if err = deployLvmd(s.cli, cluster); err != nil {
		return err
	}
	if err = initBlocks(s.cli, cluster); err != nil {
		return err
	}
	if err = deployLvmCSI(s.cli, cluster); err != nil {
		return err
	}
	return nil
}

func (s *Lvm) Update(dels, adds storagev1.Cluster) error {
	common.UpdateClusterStatusPhase(s.cli, adds.Name, storagev1.Updating)
	var err error
	defer func() {
		if err != nil {
			common.UpdateClusterStatusPhase(s.cli, adds.Name, storagev1.Failed)
		} else {
			common.UpdateClusterStatusPhase(s.cli, adds.Name, storagev1.Running)
		}
	}()
	if err = doAddhost(s.cli, adds); err != nil {
		return err
	}
	if err = doDelhost(s.cli, dels); err != nil {
		return err
	}
	return nil
}

func (s *Lvm) Delete(cluster storagev1.Cluster) error {
	common.UpdateClusterStatusPhase(s.cli, cluster.Name, storagev1.Deleting)
	var err error
	defer func() {
		if err != nil {
			common.UpdateClusterStatusPhase(s.cli, cluster.Name, storagev1.Failed)
		}
	}()
	if err = undeployLvmCSI(s.cli, cluster); err != nil {
		return err
	}
	if err = unInitBlocks(s.cli, cluster); err != nil {
		return err
	}
	if err = undeployLvmd(s.cli, cluster); err != nil {
		return err
	}
	if err = common.DeleteNodeAnnotationsAndLabels(s.cli, common.StorageHostLabels, s.GetType(), cluster.Spec.Hosts); err != nil {
		return err
	}
	if err = common.DelFinalizerForStorage(s.cli, cluster.Name, common.StoragePrestopHookFinalizer); err != nil {
		return err
	}
	return nil
}
