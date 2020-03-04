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

func undeployLvmCSI(cli client.Client, cluster storagev1.Cluster) error {
	log.Debugf("Undeploy CSI for storage cluster:%s", cluster.Spec.StorageType)
	yaml, err := csiyaml(cluster.Name)
	if err != nil {
		return err
	}
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	common.WaitStsTerminated(cli, common.StorageNamespace, CSIProvisionerStsName)
	common.WaitDsTerminated(cli, common.StorageNamespace, CSIPluginDsName)

	log.Debugf("Undeploy storageclass %s", cluster.Name)
	yaml, err = scyaml(cluster.Name)
	if err != nil {
		return err
	}
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	return nil
}

func undeployLvmd(cli client.Client, cluster storagev1.Cluster) error {
	log.Debugf("Undeploy Lvmd for storage cluster:%s", cluster.Spec.StorageType)
	yaml, err := lvmdyaml()
	if err != nil {
		return err
	}
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	common.WaitDsTerminated(cli, common.StorageNamespace, LvmdDsName)
	return nil
}

func unInitBlocks(cli client.Client, cluster storagev1.Cluster) error {
	ctx := context.TODO()
	for _, host := range cluster.Status.Config {
		if len(host.BlockDevices) == 0 {
			return fmt.Errorf("No block device to uninit for host %s", host.NodeName)
		}
		lvmdcli, err := common.CreateLvmdClient(ctx, cli, host.NodeName)
		if err != nil {
			return fmt.Errorf("Create Lvmd client failed for host %s, %v", host.NodeName, err)
		}
		defer lvmdcli.Close()
		for _, block := range host.BlockDevices {
			log.Debugf("[%s] Remove vg with %s", host.NodeName, block)
			if err := common.RemoveVG(ctx, lvmdcli, VolumeGroup); err != nil {
				return fmt.Errorf("Remove vg failed, %v", err)
			}
			log.Debugf("[%s] Remove pv with %s", host.NodeName, block)
			if err := common.RemovePV(ctx, lvmdcli, block); err != nil {
				return fmt.Errorf("Remove pv failed, %v", err)
			}
		}
	}
	return nil
}
