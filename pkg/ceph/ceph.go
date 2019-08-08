package ceph

import (
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/immense/pkg/common"
)

type Ceph struct {
	cli client.Client
}

func New(c client.Client) *Ceph {
	return &Ceph{
		cli: c,
	}
}

func (s *Ceph) GetType() string {
	return "ceph"
}

func (s *Ceph) Create(cluster common.Storage) error {
	if err := common.UpdateStatusPhase(s.cli, cluster.Name, "Creating"); err != nil {
		log.Warnf("Update storage cluster %s status failed. Err: %s", cluster.Name, err.Error())
	}
	if err := common.CreateNodeAnnotationsAndLabels(s.cli, cluster); err != nil {
		return err
	}
	if err := create(s.cli, cluster); err != nil {
		return err
	}
	if err := common.UpdateStatusPhase(s.cli, cluster.Name, "Running"); err != nil {
		log.Warnf("Update storage cluster %s status failed. Err: %s", cluster.Name, err.Error())
	}
	return nil
}

func (s *Ceph) Update(dels, adds common.Storage) error {
	if err := common.UpdateStatusPhase(s.cli, adds.Name, "Updating"); err != nil {
		log.Warnf("Update storage cluster %s status failed. Err: %s", adds.Name, err.Error())
	}
	if err := common.DeleteNodeAnnotationsAndLabels(s.cli, dels); err != nil {
		return err
	}
	if err := common.CreateNodeAnnotationsAndLabels(s.cli, adds); err != nil {
		return err
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

func (s *Ceph) Delete(cluster common.Storage) error {
	if err := delete(s.cli, cluster); err != nil {
		return err
	}
	return common.DeleteNodeAnnotationsAndLabels(s.cli, cluster)
}
