package lvm

import (
	"context"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/immense/pkg/common"
)

func doDelhost(cli client.Client, cfg map[string][]string) error {
	if len(cfg) == 0 {
		return nil
	}
	cluster := common.MakeClusterCfg(cfg, StorageType)
	log.Debugf("Delete host for storage cluster:%s, Cfg: %s", cluster.Spec.StorageType, cfg)
	if err := unInitBlocks(cli, cluster); err != nil {
		return err
	}
	return common.DeleteNodeAnnotationsAndLabels(cli, cluster)
}

func doAddhost(cli client.Client, cfg map[string][]string) error {
	if len(cfg) == 0 {
		return nil
	}
	cluster := common.MakeClusterCfg(cfg, StorageType)
	log.Debugf("Add host for storage cluster:%s, Cfg: %s", cluster.Spec.StorageType, cfg)
	if err := common.CreateNodeAnnotationsAndLabels(cli, cluster); err != nil {
		return err
	}
	return initBlocks(cli, cluster)
}

func doChangeAdd(cli client.Client, cfg map[string][]string) error {
	if len(cfg) == 0 {
		return nil
	}
	cluster := common.MakeClusterCfg(cfg, StorageType)
	log.Debugf("Add host config for storage cluster:%s, Cfg: %s", cluster.Spec.StorageType, cfg)
	if err := initBlocks(cli, cluster); err != nil {
		return err
	}
	return common.UpdateNodeAnnotations(cli, cluster)
}

func doChangeDel(cli client.Client, cfg map[string][]string) error {
	if len(cfg) == 0 {
		return nil
	}
	cluster := common.MakeClusterCfg(cfg, StorageType)
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
			if err := common.VgReduce(ctx, lvmdcli, block, VGName); err != nil {
				return err
			}
		}
	}
	return common.UpdateNodeAnnotations(cli, cluster)
}
