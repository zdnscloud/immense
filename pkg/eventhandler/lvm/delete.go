package lvm

import (
	"context"
	"errors"
	"fmt"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/eventhandler/common"
	lvmdclient "github.com/zdnscloud/lvmd/client"
)

func Delete(cli client.Client, cluster *storagev1.Cluster) error {
	if err := undeployLvmCSI(cli, cluster); err != nil {
		return err
	}
	/*
		if err := formatted(cli, cluster); err != nil {
			return err
		}*/
	if err := undeployLvmd(cli, cluster); err != nil {
		return err
	}
	return common.DeleteNodeAnnotationsAndLabels(cli, cluster, NodeLabelValue)
}

func undeployLvmCSI(cli client.Client, cluster *storagev1.Cluster) error {
	log.Debugf("Undeploy for storage cluster:%s", cluster.Spec.StorageType)
	yaml, err := csiyaml()
	if err != nil {
		return err
	}
	return helper.DeleteResourceFromYaml(cli, yaml)
}

func undeployLvmd(cli client.Client, cluster *storagev1.Cluster) error {
	log.Debugf("Undeploy for storage cluster:%s", cluster.Spec.StorageType)
	yaml, err := lvmdyaml()
	if err != nil {
		return err
	}
	return helper.DeleteResourceFromYaml(cli, yaml)
}

func formatted(cli client.Client, cluster *storagev1.Cluster) error {
	ctx := context.TODO()
	for _, host := range cluster.Spec.Hosts {
		hostip, err := common.GetHostAddr(ctx, cli, host.NodeName)
		if err != nil {
			log.Warnf("Get address from host %s failed:%s", host.NodeName, err.Error())
			return err
		}
		addr := hostip + ":" + LvmdPort

		log.Debugf("Check %s Lvmd Rinning", host.NodeName)
		if !common.WaitLvmd(addr) {
			log.Warnf("Lvmd wait run timeout:%s", addr)
			return errors.New("Lvmd not ready!" + addr)
		}

		lvmdcli, err := lvmdclient.New(addr, ConLvmdTimeout)
		defer lvmdcli.Close()
		if err != nil {
			log.Warnf("Create Lvmd client failed:%s", err.Error())
			return err
		}
		for _, block := range host.BlockDevices {
			log.Debugf("Destory block start: %s", block)
			out, err := common.Destory(ctx, lvmdcli, block)
			if err != nil {
				log.Warnf("Destory block failed:%s", err.Error())
				return nil
			}
			fmt.Println(out)
		}
	}
	return nil
}
