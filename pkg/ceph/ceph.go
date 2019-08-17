package ceph

import (
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
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

func (s *Ceph) Create(cluster storagev1.Cluster) error {
	common.UpdateStatusPhase(s.cli, cluster.Name, "Creating")
	common.CreateNodeAnnotationsAndLabels(s.cli, cluster)
	if err := create(s.cli, cluster); err != nil {
		return err
	}
	common.UpdateStatusPhase(s.cli, cluster.Name, "Running")
	return nil
}

func (s *Ceph) Update(dels, adds storagev1.Cluster) error {
	common.UpdateStatusPhase(s.cli, adds.Name, "Updating")
	common.DeleteNodeAnnotationsAndLabels(s.cli, dels)
	common.CreateNodeAnnotationsAndLabels(s.cli, adds)
	if err := doAddhost(s.cli, adds); err != nil {
		return err
	}
	if err := doDelhost(s.cli, dels); err != nil {
		return err
	}
	common.UpdateStatusPhase(s.cli, adds.Name, "Running")
	return nil
}

func (s *Ceph) Delete(cluster storagev1.Cluster) error {
	common.DeleteNodeAnnotationsAndLabels(s.cli, cluster)
	return delete(s.cli, cluster)
}
