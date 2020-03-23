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
	//StorageDriverName     = "csi-lvmplugin"
	LvmDriverSuffix = "lvm.storage.zcloud.cn"
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
	go StatusControl(s.cli, cluster.Name)
	common.UpdateClusterStatusPhase(s.cli, cluster.Name, storagev1.Creating)
	if err := common.CreateNodeAnnotationsAndLabels(s.cli, common.StorageHostLabels, s.GetType(), cluster.Spec.Hosts); err != nil {
		return err
	}
	if err := deployLvmd(s.cli, cluster); err != nil {
		common.UpdateClusterStatusPhase(s.cli, cluster.Name, storagev1.Failed)
		return err
	}
	if err := initBlocks(s.cli, cluster); err != nil {
		common.UpdateClusterStatusPhase(s.cli, cluster.Name, storagev1.Failed)
		return err
	}
	if err := deployLvmCSI(s.cli, cluster); err != nil {
		common.UpdateClusterStatusPhase(s.cli, cluster.Name, storagev1.Failed)
		return err
	}

	common.UpdateClusterStatusPhase(s.cli, cluster.Name, storagev1.Running)
	return common.AddFinalizerForStorage(s.cli, cluster.Name, common.StoragePrestopHookFinalizer)
}

func (s *Lvm) Update(dels, adds storagev1.Cluster) error {
	common.UpdateClusterStatusPhase(s.cli, adds.Name, storagev1.Updating)
	if err := doAddhost(s.cli, adds); err != nil {
		common.UpdateClusterStatusPhase(s.cli, adds.Name, storagev1.Failed)
		return err
	}
	if err := doDelhost(s.cli, dels); err != nil {
		common.UpdateClusterStatusPhase(s.cli, adds.Name, storagev1.Failed)
		return err
	}
	common.UpdateClusterStatusPhase(s.cli, adds.Name, storagev1.Running)
	return nil
}

func (s *Lvm) Delete(cluster storagev1.Cluster) error {
	common.UpdateClusterStatusPhase(s.cli, cluster.Name, storagev1.Deleting)
	if err := undeployLvmCSI(s.cli, cluster); err != nil {
		common.UpdateClusterStatusPhase(s.cli, cluster.Name, storagev1.Failed)
		return err
	}
	if err := unInitBlocks(s.cli, cluster); err != nil {
		common.UpdateClusterStatusPhase(s.cli, cluster.Name, storagev1.Failed)
		return err
	}
	if err := undeployLvmd(s.cli, cluster); err != nil {
		common.UpdateClusterStatusPhase(s.cli, cluster.Name, storagev1.Failed)
		return err
	}
	if err := common.DeleteNodeAnnotationsAndLabels(s.cli, common.StorageHostLabels, s.GetType(), cluster.Spec.Hosts); err != nil {
		return err
	}
	return common.DelFinalizerForStorage(s.cli, cluster.Name, common.StoragePrestopHookFinalizer)
}
