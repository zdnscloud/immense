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
	StorageType = "lvm"
	VGName      = "k8s"
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
	common.UpdateStatusPhase(s.cli, cluster.Name, "Creating")
	common.CreateNodeAnnotationsAndLabels(s.cli, cluster)
	if err := deployLvmd(s.cli, cluster); err != nil {
		return err
	}
	if err := initBlocks(s.cli, cluster); err != nil {
		return err
	}
	if err := deployLvmCSI(s.cli, cluster); err != nil {
		return err
	}
	common.UpdateStatusPhase(s.cli, cluster.Name, "Running")
	go StatusControl(s.cli, cluster.Name)
	return nil
}

func (s *Lvm) Update(dels, adds storagev1.Cluster) error {
	common.UpdateStatusPhase(s.cli, adds.Name, "Updating")
	if err := doAddhost(s.cli, adds); err != nil {
		return err
	}
	if err := doDelhost(s.cli, dels); err != nil {
		return err
	}
	common.UpdateStatusPhase(s.cli, adds.Name, "Running")
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
