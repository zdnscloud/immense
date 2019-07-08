package lvm

import (
	"context"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/common"
)

func deployLvmCSI(cli client.Client, cluster *storagev1.Cluster) error {
	log.Debugf("Deploy CSI for storage cluster: %s", cluster.Spec.StorageType)
	yaml, err := csiyaml()
	if err != nil {
		return err
	}
	return helper.CreateResourceFromYaml(cli, yaml)
}

func deployLvmd(cli client.Client, cluster *storagev1.Cluster) error {
	log.Debugf("Deploy Lvmd for storage cluster: %s", cluster.Spec.StorageType)
	yaml, err := lvmdyaml()
	if err != nil {
		return err
	}
	return helper.CreateResourceFromYaml(cli, yaml)
}

func initBlocks(cli client.Client, cluster *storagev1.Cluster) error {
	ctx := context.TODO()
	for _, host := range cluster.Spec.Hosts {
		lvmdcli, err := common.CreateLvmdClient(ctx, cli, host.NodeName)
		if err != nil {
			log.Warnf("[%s] Create Lvmd client failed. Err: %s. Skip it", host.NodeName, err.Error())
			continue
		}
		defer lvmdcli.Close()
		for _, block := range host.BlockDevices {
			log.Debugf("[%s] Init block start: %s", host.NodeName, block)
			name, err := common.GetVG(ctx, lvmdcli, block)
			if err != nil {
				log.Warnf("Get VGName failed:%s", err.Error())
				return err
			}
			if name == VGName {
				log.Debugf("[%s] Block had inited before, skip %s", host.NodeName, block)
				continue
			}
			log.Debugf("[%s] Validate block %s", host.NodeName, block)
			if err := common.Validate(ctx, lvmdcli, block); err != nil {
				return err
			}
			log.Debugf("[%s] Create pv with %s", host.NodeName, block)
			if err := common.CreatePV(ctx, lvmdcli, block); err != nil {
				return err
			}
			log.Debugf("[%s] Create vg with %s", host.NodeName, block)
			if err := common.CreateVG(ctx, lvmdcli, block, VGName); err != nil {
				return err
			}
		}
	}
	return nil
}
