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
	StorageType      = "lvm"
	VGName           = "k8s"
	StorageClassName = "lvm"
)

func New(c client.Client) *Lvm {
	return &Lvm{
		cli: c,
	}
}

func (s *Lvm) GetType() string {
	return StorageType
}

func (s *Lvm) Create(cluster *storagev1.Cluster) error {
	if err := common.CreateNodeAnnotationsAndLabels(s.cli, cluster); err != nil {
		return err
	}
	if err := deployLvmd(s.cli, cluster); err != nil {
		return err
	}
	if err := initBlocks(s.cli, cluster); err != nil {
		return err
	}
	if err := deployLvmCSI(s.cli, cluster); err != nil {
		return err
	}
	go StatusControl(s.cli, cluster.Name)
	return nil
}

func (s *Lvm) Update(oldcfg, newcfg *storagev1.Cluster) error {
	delcfg, addcfg, changetodel, changetoadd := common.Diff(oldcfg, newcfg)
	if err := doAddhost(s.cli, addcfg); err != nil {
		return err
	}
	if err := doChangeAdd(s.cli, changetoadd); err != nil {
		return err
	}
	if err := doDelhost(s.cli, delcfg); err != nil {
		return err
	}
	if err := doChangeDel(s.cli, changetodel); err != nil {
		return err
	}
	return nil
}

func (s *Lvm) Delete(cluster *storagev1.Cluster) error {
	if err := undeployLvmCSI(s.cli, cluster); err != nil {
		return err
	}
	if err := unInitBlocks(s.cli, cluster); err != nil {
		return err
	}
	if err := undeployLvmd(s.cli, cluster); err != nil {
		return err
	}
	return common.DeleteNodeAnnotationsAndLabels(s.cli, cluster)
}
