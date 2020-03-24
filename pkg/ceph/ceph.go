package ceph

import (
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/ceph/status"
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
	common.UpdateClusterStatusPhase(s.cli, cluster.Name, storagev1.Creating)
	go status.Watch(s.cli, cluster.Name)
	var err error
	defer func() {
		if err != nil {
			common.UpdateClusterStatusPhase(s.cli, cluster.Name, storagev1.Failed)
		} else {
			common.UpdateClusterStatusPhase(s.cli, cluster.Name, storagev1.Running)
		}
	}()
	if err = common.AddFinalizerForStorage(s.cli, cluster.Name, common.StoragePrestopHookFinalizer); err != nil {
		return err
	}
	if err = common.CreateNodeAnnotationsAndLabels(s.cli, common.StorageHostLabels, s.GetType(), cluster.Spec.Hosts); err != nil {
		return err
	}
	if err = create(s.cli, cluster); err != nil {
		return err
	}
	return nil
}

func (s *Ceph) Update(dels, adds storagev1.Cluster) error {
	common.UpdateClusterStatusPhase(s.cli, adds.Name, storagev1.Updating)
	var err error
	defer func() {
		if err != nil {
			common.UpdateClusterStatusPhase(s.cli, adds.Name, storagev1.Failed)
		} else {
			common.UpdateClusterStatusPhase(s.cli, adds.Name, storagev1.Running)
		}
	}()
	if err = common.DeleteNodeAnnotationsAndLabels(s.cli, common.StorageHostLabels, s.GetType(), dels.Spec.Hosts); err != nil {
		return err
	}
	if err = common.CreateNodeAnnotationsAndLabels(s.cli, common.StorageHostLabels, s.GetType(), adds.Spec.Hosts); err != nil {
		return err
	}
	if err = doAddhost(s.cli, adds); err != nil {
		return err
	}
	if err = doDelhost(s.cli, dels); err != nil {
		return err
	}
	if err = updatePgNumIfNeed(s.cli, adds.Name); err != nil {
		return err
	}
	return nil
}

func (s *Ceph) Delete(cluster storagev1.Cluster) error {
	common.UpdateClusterStatusPhase(s.cli, cluster.Name, storagev1.Deleting)
	var err error
	defer func() {
		if err != nil {
			common.UpdateClusterStatusPhase(s.cli, cluster.Name, storagev1.Failed)
		}
	}()
	if err = common.DeleteNodeAnnotationsAndLabels(s.cli, common.StorageHostLabels, s.GetType(), cluster.Spec.Hosts); err != nil {
		return err
	}
	if err = delete(s.cli, cluster); err != nil {
		return err
	}
	if err = common.DelFinalizerForStorage(s.cli, cluster.Name, common.StoragePrestopHookFinalizer); err != nil {
		return err
	}
	return nil
}
