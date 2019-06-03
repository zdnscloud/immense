package lvm

import (
	"context"
	"errors"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/eventhandler/common"
	lvmdclient "github.com/zdnscloud/lvmd/client"
	"time"
)

const (
	NodeLabelValue = "Lvm"
	LvmdPort       = "1736"
)

var ConLvmdTimeout = 5 * time.Second

func Create(cli client.Client, cluster *storagev1.Cluster) error {
	if err := common.CreateNodeAnnotationsAndLabels(cli, cluster, NodeLabelValue); err != nil {
		return err
	}
	if err := deployLvmd(cli, cluster); err != nil {
		return err
	}
	if err := initBlocks(cli, cluster); err != nil {
		return err
	}
	return nil
	return deployLvmCSI(cli, cluster)
}

func deployLvmCSI(cli client.Client, cluster *storagev1.Cluster) error {
	log.Debugf("Deploy CSI for storage cluster: %s", cluster.Spec.StorageType)
	cfg := map[string]interface{}{
		"AoDNamespace":                   "yes",
		"RBACConfig":                     common.RBACConfig,
		"LabelKey":                       common.StorageHostLabels,
		"LabelValue":                     NodeLabelValue,
		"StorageNamespace":               common.StorageNamespace,
		"StorageLvmAttacherImage":        "quay.io/k8scsi/csi-attacher:v1.0.0",
		"StorageLvmProvisionerImage":     "quay.io/k8scsi/csi-provisioner:v1.0.0",
		"StorageLvmDriverRegistrarImage": "quay.io/k8scsi/csi-node-driver-registrar:v1.0.2",
		"StorageLvmCSIImage":             "zdnscloud/lvmcsi:v0.5",
		"StorageClassName":               "lvm",
	}
	yaml, err := common.CompileTemplateFromMap(LvmCSITemplate, cfg)
	if err != nil {
		return err
	}
	return helper.CreateResourceFromYaml(cli, yaml)
}

func deployLvmd(cli client.Client, cluster *storagev1.Cluster) error {
	log.Debugf("Deploy LvmD for storage cluster: %s", cluster.Spec.StorageType)
	cfg := map[string]interface{}{
		"AoDNamespace":     "yes",
		"RBACConfig":       common.RBACConfig,
		"LabelKey":         common.StorageHostLabels,
		"LabelValue":       NodeLabelValue,
		"StorageNamespace": common.StorageNamespace,
		"StorageLvmdImage": "zdnscloud/lvmd:v0.94",
	}
	yaml, err := common.CompileTemplateFromMap(LvmDTemplate, cfg)
	if err != nil {
		return err
	}
	return helper.CreateResourceFromYaml(cli, yaml)
}

func initBlocks(cli client.Client, cluster *storagev1.Cluster) error {
	ctx := context.TODO()
	for _, host := range cluster.Spec.Hosts {
		hostip, err := common.GetHostAddr(ctx, cli, host.NodeName)
		if err != nil {
			return err
		}
		addr := hostip + ":" + LvmdPort

		log.Debugf("Check %s Lvmd Rinning", host.NodeName)
		if !common.WaitLvmd(addr) {
			return errors.New("Lvmd not ready!" + addr)
		}

		lvmdcli, err := lvmdclient.New(addr, ConLvmdTimeout)
		defer lvmdcli.Close()
		if err != nil {
			return err
		}

		for _, block := range host.BlockDevices {
			log.Debugf("Init block: %s", block)
			name, err := common.GetVG(ctx, lvmdcli, block)
			if err != nil {
				log.Warnf("Get VGName failed:%s", err.Error())
				return nil
			}
			if name == common.VGName {
				log.Debugf("Block had inited before")
				continue
			}

			v, err := common.Validate(ctx, lvmdcli, block)
			if err != nil {
				log.Warnf("Validate block failed:%s", err.Error())
				return err
			}
			if !v {
				return errors.New("Some blocks cat not be used!" + block)
			}

			p, err := common.PvExist(ctx, lvmdcli, block)
			if err != nil {
				log.Warnf("Check PV failed:%s", err.Error())
				return err
			}
			if !p {
				if err := common.PvCreate(ctx, lvmdcli, block); err != nil {
					log.Warnf("Create PV failed:%s", err.Error())
					return err
				}
			}

			v, err = common.VgExist(ctx, lvmdcli)
			if err != nil {
				log.Warnf("Check VG failed:%s", err.Error())
				return err
			}
			if v {
				if common.VgExtend(ctx, lvmdcli, block); err != nil {
					log.Warnf("Vgextend block failed:%s", err.Error())
					return err
				}
				continue
			}

			if err := common.VgCreate(ctx, lvmdcli, block); err != nil {
				log.Warnf("Create VG failed:%s", err.Error())
				return err
			}
		}
	}
	return nil
}
