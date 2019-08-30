package lvm

import (
	"context"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/lvm/util"
)

func undeployLvmCSI(cli client.Client, cluster storagev1.Cluster) error {
	log.Debugf("Undeploy CSI for storage cluster:%s", cluster.Spec.StorageType)
	yaml, err := csiyaml(cluster.Name)
	if err != nil {
		return err
	}
	return helper.DeleteResourceFromYaml(cli, yaml)
}

func undeployLvmd(cli client.Client, cluster storagev1.Cluster) error {
	log.Debugf("Undeploy Lvmd for storage cluster:%s", cluster.Spec.StorageType)
	yaml, err := lvmdyaml()
	if err != nil {
		return err
	}
	return helper.DeleteResourceFromYaml(cli, yaml)
}

func unInitBlocks(cli client.Client, cluster storagev1.Cluster) error {
	ctx := context.TODO()
	for _, host := range cluster.Status.Config {
		if len(host.BlockDevices) == 0 {
			log.Debugf("[%s] No block device to uninit", host.NodeName)
			continue
		}
		lvmdcli, err := util.CreateLvmdClient(ctx, cli, host.NodeName)
		if err != nil {
			log.Warnf("[%s] Create Lvmd client failed:%s", host.NodeName, err.Error())
			continue
		}
		defer lvmdcli.Close()
		for _, block := range host.BlockDevices {
			log.Debugf("[%s] Remove vg with %s", host.NodeName, block.Name)
			if err := util.RemoveVG(ctx, lvmdcli, VGName); err != nil {
				log.Warnf("[%s] Remove vg with %s failed:%s", host.NodeName, block.Name, err.Error())
				continue
			}
			log.Debugf("[%s] Remove pv with %s", host.NodeName, block)
			if err := util.RemovePV(ctx, lvmdcli, block.Name); err != nil {
				log.Warnf("[%s] Remove pv with %s failed:%s", host.NodeName, block.Name, err.Error())
				continue
			}
		}
	}
	return nil
}
