package lvm

import (
	"context"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/common"
	"time"
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
	common.WaitCSIReady(cli, CSIProvisionerStsName, CSIPluginDsName)

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
	waitDone(cli)
	return nil
}

func initBlocks(cli client.Client, cluster storagev1.Cluster) error {
	ctx := context.TODO()
	for _, host := range cluster.Status.Config {
		if len(host.BlockDevices) == 0 {
			log.Debugf("[%s] No block device to init", host.NodeName)
			continue
		}
		lvmdcli, err := CreateLvmdClient(ctx, cli, host.NodeName)
		if err != nil {
			log.Warnf("[%s] Create Lvmd client failed. Err: %s. Skip it", host.NodeName, err.Error())
			continue
		}
		defer lvmdcli.Close()
		for _, block := range host.BlockDevices {
			log.Debugf("[%s] Init block start: %s", host.NodeName, block)
			name, err := GetVG(ctx, lvmdcli, block)
			if err != nil {
				log.Warnf("Get VGName failed:%s", err.Error())
				return err
			}
			if name == VGName {
				log.Debugf("[%s] Block had inited before, skip %s", host.NodeName, block)
				continue
			}
			log.Debugf("[%s] Validate block %s", host.NodeName, block)
			if err := Validate(ctx, lvmdcli, block); err != nil {
				log.Warnf("[%s] Validate block %s failed:%s", host.NodeName, block, err.Error())
				continue
			}
			log.Debugf("[%s] Create pv with %s", host.NodeName, block)
			if err := CreatePV(ctx, lvmdcli, block); err != nil {
				log.Warnf("[%s] Create pv with %s failed:%s", host.NodeName, block, err.Error())
				continue
			}
			log.Debugf("[%s] Create vg with %s", host.NodeName, block)
			if err := CreateVG(ctx, lvmdcli, block, VGName); err != nil {
				log.Warnf("[%s] Create vg with %s failed:%s", host.NodeName, block, err.Error())
				continue
			}
		}
	}
	return nil
}

func waitDone(cli client.Client) {
	log.Debugf("Wait all lvmd running, this will take some time")
	var done bool
	for !done {
		time.Sleep(10 * time.Second)
		if !common.IsDsReady(cli, common.StorageNamespace, LvmdDsName) {
			continue
		}
		done = true
	}
}
