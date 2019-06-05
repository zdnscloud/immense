package lvm

import (
	"context"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/eventhandler/common"
)

func Update(cli client.Client, oldcfg, newcfg *storagev1.Cluster) error {
	delcfg, addcfg, changetodel, changetoadd := common.Diff(oldcfg, newcfg)
	if err := doDelhost(cli, delcfg); err != nil {
		return err
	}
	if err := doAddhost(cli, addcfg); err != nil {
		return err
	}
	if err := doChangeAdd(cli, changetoadd); err != nil {
		return err
	}
	if err := doChangeDel(cli, changetodel); err != nil {
		return err
	}
	return nil
}

func doDelhost(cli client.Client, cfg map[string][]string) error {
	if len(cfg) == 0 {
		return nil
	}
	cluster := makeClusterCfg(cfg)
	log.Debugf("Delete host for storage cluster:%s, Cfg: %s", cluster.Spec.StorageType, cfg)
	if err := unInitBlocks(cli, cluster); err != nil {
		return err
	}
	if err := common.DeleteNodeAnnotationsAndLabels(cli, cluster, NodeLabelValue); err != nil {
		return err
	}
	return nil
}

func doAddhost(cli client.Client, cfg map[string][]string) error {
	if len(cfg) == 0 {
		return nil
	}
	cluster := makeClusterCfg(cfg)
	log.Debugf("Add host for storage cluster:%s, Cfg: %s", cluster.Spec.StorageType, cfg)
	if err := common.CreateNodeAnnotationsAndLabels(cli, cluster, NodeLabelValue); err != nil {
		return err
	}
	if err := initBlocks(cli, cluster); err != nil {
		return err
	}
	return nil
}

func doChangeAdd(cli client.Client, cfg map[string][]string) error {
	if len(cfg) == 0 {
		return nil
	}
	cluster := makeClusterCfg(cfg)
	log.Debugf("Add host config for storage cluster:%s, Cfg: %s", cluster.Spec.StorageType, cfg)
	if err := initBlocks(cli, cluster); err != nil {
		return err
	}
	return nil
}

func doChangeDel(cli client.Client, cfg map[string][]string) error {
	if len(cfg) == 0 {
		return nil
	}
	cluster := makeClusterCfg(cfg)
	log.Debugf("Delete host config for storage cluster:%s, Cfg: %s", cluster.Spec.StorageType, cfg)
	ctx := context.TODO()
	for _, host := range cluster.Spec.Hosts {
		lvmdcli, err := common.CreateLvmdClient(ctx, cli, host.NodeName)
		if err != nil {
			log.Warnf("[%s] Create Lvmd client failed:%s", host.NodeName, err.Error())
			return err
		}
		defer lvmdcli.Close()
		for _, block := range host.BlockDevices {
			log.Debugf("[%s] Reduce vg and Remove pv with %s", host.NodeName, block)
			if err := common.VgReduce(ctx, lvmdcli, block); err != nil {
				return nil
			}
		}
	}
	return nil
}

func makeClusterCfg(cfg map[string][]string) *storagev1.Cluster {
	hosts := make([]storagev1.HostSpec, 0)

	for k, v := range cfg {
		host := storagev1.HostSpec{
			NodeName:     k,
			BlockDevices: v,
		}
		hosts = append(hosts, host)
	}
	return &storagev1.Cluster{
		Spec: storagev1.ClusterSpec{
			StorageType: NodeLabelValue,
			Hosts:       hosts,
		},
	}
}
