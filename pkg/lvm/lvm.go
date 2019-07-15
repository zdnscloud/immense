package lvm

import (
	"github.com/zdnscloud/cement/log"
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
	status := storagev1.ClusterStatus{
		Phase: "Creating"}
	if err := common.UpdateStatus(s.cli, cluster.Name, status); err != nil {
		log.Warnf("Update storage cluster %s status failed. Err: %s", cluster.Name, err.Error())
	}
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
	status = storagev1.ClusterStatus{
		Phase: "Running"}
	if err := common.UpdateStatus(s.cli, cluster.Name, status); err != nil {
		log.Warnf("Update storage cluster %s status failed. Err: %s", cluster.Name, err.Error())
	}
	return nil
}

func (s *Lvm) Update(oldcfg, newcfg *storagev1.Cluster) error {
	status := storagev1.ClusterStatus{
		Phase: "Updating"}
	if err := common.UpdateStatus(s.cli, newcfg.Name, status); err != nil {
		log.Warnf("Update storage cluster %s status failed. Err: %s", newcfg.Name, err.Error())
	}
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
