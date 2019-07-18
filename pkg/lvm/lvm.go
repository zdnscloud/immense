package lvm

import (
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
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

func (s *Lvm) Create(cluster common.Storage) error {
	if err := common.UpdateStatusPhase(s.cli, cluster.Name, "Creating"); err != nil {
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
	if err := common.UpdateStatusPhase(s.cli, cluster.Name, "Running"); err != nil {
		log.Warnf("Update storage cluster %s status failed. Err: %s", cluster.Name, err.Error())
	}
	go StatusControl(s.cli, cluster.Name)
	return nil
}

func (s *Lvm) Update(dels, adds common.Storage) error {
	if err := common.UpdateStatusPhase(s.cli, adds.Name, "Updating"); err != nil {
		log.Warnf("Update storage cluster %s status failed. Err: %s", adds.Name, err.Error())
	}
	if err := doAddhost(s.cli, adds); err != nil {
		return err
	}
	if err := doDelhost(s.cli, dels); err != nil {
		return err
	}
	if err := common.UpdateStatusPhase(s.cli, adds.Name, "Running"); err != nil {
		log.Warnf("Update storage cluster %s status failed. Err: %s", adds.Name, err.Error())
	}
	return nil
}

func (s *Lvm) Delete(cluster common.Storage) error {
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
