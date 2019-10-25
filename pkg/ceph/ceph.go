package ceph

import (
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/ceph/global"
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
	return global.StorageType
}

func (s *Ceph) Create(cluster storagev1.Cluster) error {
	common.UpdateStatusPhase(s.cli, cluster.Name, common.Creating)
	common.CreateNodeAnnotationsAndLabels(s.cli, cluster)
	if err := create(s.cli, cluster); err != nil {
		common.UpdateStatusPhase(s.cli, cluster.Name, common.Failed)
		return err
	}
	common.UpdateStatusPhase(s.cli, cluster.Name, common.Running)
	return nil
}

func (s *Ceph) Update(dels, adds storagev1.Cluster) error {
	common.UpdateStatusPhase(s.cli, adds.Name, common.Updating)
	common.DeleteNodeAnnotationsAndLabels(s.cli, dels)
	common.CreateNodeAnnotationsAndLabels(s.cli, adds)
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

func (s *Ceph) Delete(cluster storagev1.Cluster) error {
	common.DeleteNodeAnnotationsAndLabels(s.cli, cluster)
	return delete(s.cli, cluster)
}
