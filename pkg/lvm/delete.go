package lvm

import (
	"context"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/common"
)

func undeployLvmCSI(cli client.Client, cluster *storagev1.Cluster) error {
	log.Debugf("Undeploy CSI for storage cluster:%s", cluster.Spec.StorageType)
	yaml, err := csiyaml("keepns")
	if err != nil {
		return err
	}
	return helper.DeleteResourceFromYaml(cli, yaml)
}

func undeployLvmd(cli client.Client, cluster *storagev1.Cluster) error {
	log.Debugf("Undeploy Lvmd for storage cluster:%s", cluster.Spec.StorageType)
	yaml, err := lvmdyaml("keepns")
	if err != nil {
		return err
	}
	return helper.DeleteResourceFromYaml(cli, yaml)
}

func unInitBlocks(cli client.Client, cluster *storagev1.Cluster) error {
	ctx := context.TODO()
	for _, host := range cluster.Spec.Hosts {
		lvmdcli, err := common.CreateLvmdClient(ctx, cli, host.NodeName)
		if err != nil {
			log.Warnf("[%s] Create Lvmd client failed:%s", host.NodeName, err.Error())
			return err
		}
		defer lvmdcli.Close()
		for _, block := range host.BlockDevices {
			log.Debugf("[%s] Remove vg with %s", host.NodeName, block)
			if err := common.RemoveVG(ctx, lvmdcli, VGName); err != nil {
				return err
			}
			log.Debugf("[%s] Remove pv with %s", host.NodeName, block)
			if err := common.RemovePV(ctx, lvmdcli, block); err != nil {
				return err
			}
		}
	}
	return nil
}
