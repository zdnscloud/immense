package lvm

import (
	"context"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/lvm/util"
)

func deployLvmCSI(cli client.Client, cluster storagev1.Cluster) error {
	log.Debugf("Deploy CSI for storage cluster: %s", cluster.Spec.StorageType)
	yaml, err := csiyaml(cluster.Name)
	if err != nil {
		return err
	}
	return helper.CreateResourceFromYaml(cli, yaml)
}

func deployLvmd(cli client.Client, cluster storagev1.Cluster) error {
	log.Debugf("Deploy Lvmd for storage cluster: %s", cluster.Spec.StorageType)
	yaml, err := lvmdyaml()
	if err != nil {
		return err
	}
	return helper.CreateResourceFromYaml(cli, yaml)
}

func initBlocks(cli client.Client, cluster storagev1.Cluster) error {
	ctx := context.TODO()
	for _, host := range cluster.Status.Config {
		if len(host.BlockDevices) == 0 {
			log.Debugf("[%s] No block device to init", host.NodeName)
			continue
		}
		lvmdcli, err := util.CreateLvmdClient(ctx, cli, host.NodeName)
		if err != nil {
			log.Warnf("[%s] Create Lvmd client failed. Err: %s. Skip it", host.NodeName, err.Error())
			continue
		}
		defer lvmdcli.Close()
		for _, block := range host.BlockDevices {
			log.Debugf("[%s] Init block start: %s", host.NodeName, block.Name)
			name, err := util.GetVG(ctx, lvmdcli, block.Name)
			if err != nil {
				log.Warnf("Get VGName failed:%s", err.Error())
				return err
			}
			if name == VGName {
				log.Debugf("[%s] Block had inited before, skip %s", host.NodeName, block.Name)
				continue
			}
			log.Debugf("[%s] Validate block %s", host.NodeName, block.Name)
			if err := util.Validate(ctx, lvmdcli, block.Name); err != nil {
				log.Warnf("[%s] Validate block %s failed:%s", host.NodeName, block, err.Error())
				continue
			}
			log.Debugf("[%s] Create pv with %s", host.NodeName, block.Name)
			if err := util.CreatePV(ctx, lvmdcli, block.Name); err != nil {
				log.Warnf("[%s] Create pv with %s failed:%s", host.NodeName, block, err.Error())
				continue
			}
			log.Debugf("[%s] Create vg with %s", host.NodeName, block.Name)
			if err := util.CreateVG(ctx, lvmdcli, block.Name, VGName); err != nil {
				log.Warnf("[%s] Create vg with %s failed:%s", host.NodeName, block, err.Error())
				continue
			}
		}
	}
	return nil
}
