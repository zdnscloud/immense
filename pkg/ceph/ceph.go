package ceph

import (
	"github.com/zdnscloud/cement/log"
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

func (s *Ceph) Create(cluster *storagev1.Cluster) error {
	if err := common.CreateNodeAnnotationsAndLabels(s.cli, cluster); err != nil {
		return err
	}
	return create(s.cli, cluster)
}

func (s *Ceph) Update(oldcfg, newcfg *storagev1.Cluster) error {
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
	if err := common.DeleteNodeAnnotationsAndLabels(s.cli, oldcfg); err != nil {
		return err
	}
	return common.CreateNodeAnnotationsAndLabels(s.cli, newcfg)
}

func (s *Ceph) Delete(cluster *storagev1.Cluster) error {
	if err := delete(s.cli, cluster); err != nil {
		return err
	}
	return common.DeleteNodeAnnotationsAndLabels(s.cli, cluster)
}
