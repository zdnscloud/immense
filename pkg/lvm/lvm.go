package lvm

import (
	"errors"
	"fmt"
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

func New(c client.Client) (*Lvm, error) {
	return &Lvm{
		cli: c,
	}, nil
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
	var usedHost string
	for node := range changetodel {
		used, err := CheckUsed(s.cli, node)
		if err != nil {
			return err
		}
		if used {
			usedHost = usedHost + node + ","
		}
	}
	for node := range delcfg {
		used, err := CheckUsed(s.cli, node)
		if err != nil {
			return err
		}
		if used {
			usedHost = usedHost + node + ","
		}
	}
	if len(usedHost) > 0 {
		return errors.New(fmt.Sprintf("Host %v block device is used by pod, can not to be remove", usedHost))
	}
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
	var usedHost string
	for _, node := range cluster.Spec.Hosts {
		used, err := CheckUsed(s.cli, node.NodeName)
		if err != nil {
			return err
		}
		if used {
			usedHost = usedHost + node.NodeName + ","
		}
	}
	if len(usedHost) > 0 {
		return errors.New(fmt.Sprintf("Host %v block device is used by pod, can not to be remove", usedHost))
	}
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
