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
	VGName                = "k8s"
	LvmdDsName            = "lvmd"
	CSIPluginDsName       = "csi-lvmplugin"
	CSIProvisionerStsName = "csi-lvmplugin-provisioner"
	StorageDriverName     = "csi-lvmplugin"
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
	common.UpdateStatusPhase(s.cli, cluster.Name, common.Creating)
	common.CreateNodeAnnotationsAndLabels(s.cli, cluster)
	if err := deployLvmd(s.cli, cluster); err != nil {
		common.UpdateStatusPhase(s.cli, cluster.Name, common.Failed)
		return err
	}
	if err := initBlocks(s.cli, cluster); err != nil {
		common.UpdateStatusPhase(s.cli, cluster.Name, common.Failed)
		return err
	}
	if err := deployLvmCSI(s.cli, cluster); err != nil {
		common.UpdateStatusPhase(s.cli, cluster.Name, common.Failed)
		return err
	}

	common.UpdateStatusPhase(s.cli, cluster.Name, common.Running)
	go StatusControl(s.cli, cluster.Name)
	return nil
}

func (s *Lvm) Update(dels, adds storagev1.Cluster) error {
	common.UpdateStatusPhase(s.cli, adds.Name, common.Updating)
	if err := doAddhost(s.cli, adds); err != nil {
		common.UpdateStatusPhase(s.cli, adds.Name, common.Failed)
		return err
	}
	if err := doDelhost(s.cli, dels); err != nil {
		common.UpdateStatusPhase(s.cli, adds.Name, common.Failed)
		return err
	}
	common.UpdateStatusPhase(s.cli, adds.Name, common.Running)
	return nil
}

func (s *Lvm) Delete(cluster storagev1.Cluster) error {
	if err := undeployLvmCSI(s.cli, cluster); err != nil {
		return err
	}
	if err := unInitBlocks(s.cli, cluster); err != nil {
		return err
	}
	if err := undeployLvmd(s.cli, cluster); err != nil {
		return err
	}
	common.DeleteNodeAnnotationsAndLabels(s.cli, cluster)
	return nil
}
