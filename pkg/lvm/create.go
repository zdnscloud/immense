package lvm

import (
	"context"
	"fmt"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/common"
)

func deployLvmCSI(cli client.Client, cluster storagev1.Cluster) error {
	log.Debugf("Deploy lvmcsi")
	yaml, err := csiyaml(cluster.Name)
	if err != nil {
		return err
	}
	if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	common.WaitStsReady(cli, common.StorageNamespace, CSIProvisionerStsName)
	common.WaitDsReady(cli, common.StorageNamespace, CSIPluginDsName)

	log.Debugf("Deploy storageclass %s", cluster.Name)
	yaml, err = scyaml(cluster.Name)
	if err != nil {
		return err
	}
	return helper.CreateResourceFromYaml(cli, yaml)
}

func deployLvmd(cli client.Client, cluster storagev1.Cluster) error {
	log.Debugf("Deploy Lvmd")
	yaml, err := lvmdyaml()
	if err != nil {
		return err
	}
	if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	common.WaitDsReady(cli, common.StorageNamespace, LvmdDsName)
	return nil
}

func initBlocks(cli client.Client, cluster storagev1.Cluster) error {
	ctx := context.TODO()
	for _, host := range cluster.Status.Config {
		if len(host.BlockDevices) == 0 {
			return fmt.Errorf("No block device to init for host %s", host.NodeName)
		}
		lvmdcli, err := common.CreateLvmdClient(ctx, cli, host.NodeName)
		if err != nil {
			return fmt.Errorf("Create Lvmd client failed for host %s, %v", host.NodeName, err)
		}
		defer lvmdcli.Close()
		for _, block := range host.BlockDevices {
			log.Debugf("[%s] create volumegroup for block %s", host.NodeName, block)
			if err := common.GenVolumeGroup(lvmdcli, block, VolumeGroup); err != nil {
				return err
			}
		}
	}
	return nil
}
