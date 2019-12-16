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
	common.UpdateStatusPhase(s.cli, cluster.Name, storagev1.Creating)
	common.CreateNodeAnnotationsAndLabels(s.cli, cluster)
	if err := create(s.cli, cluster); err != nil {
		common.UpdateStatusPhase(s.cli, cluster.Name, storagev1.Failed)
		return err
	}
	common.UpdateStatusPhase(s.cli, cluster.Name, storagev1.Running)
	return common.AddFinalizerForStorage(s.cli, cluster.Name, common.ClusterPrestopHookFinalizer)
}

func (s *Ceph) Update(dels, adds storagev1.Cluster) error {
	common.UpdateStatusPhase(s.cli, adds.Name, storagev1.Updating)
	common.DeleteNodeAnnotationsAndLabels(s.cli, dels)
	common.CreateNodeAnnotationsAndLabels(s.cli, adds)
	if err := doAddhost(s.cli, adds); err != nil {
		common.UpdateStatusPhase(s.cli, adds.Name, storagev1.Failed)
		return err
	}
	if err := doDelhost(s.cli, dels); err != nil {
		common.UpdateStatusPhase(s.cli, adds.Name, storagev1.Failed)
		return err
	}
	if err := updatePgNumIfNeed(s.cli, adds.Name); err != nil {
		common.UpdateStatusPhase(s.cli, adds.Name, storagev1.Failed)
		return err
	}
	common.UpdateStatusPhase(s.cli, adds.Name, storagev1.Running)
	return nil
}

func (s *Ceph) Delete(cluster storagev1.Cluster) error {
	common.UpdateStatusPhase(s.cli, cluster.Name, storagev1.Deleting)
	common.DeleteNodeAnnotationsAndLabels(s.cli, cluster)
	if err := delete(s.cli, cluster); err != nil {
		common.UpdateStatusPhase(s.cli, cluster.Name, storagev1.Failed)
		return err
	}
	return common.DelFinalizerForStorage(s.cli, cluster.Name, common.ClusterPrestopHookFinalizer)
}
