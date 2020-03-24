package iscsi

import (
	"errors"
	"fmt"
	"time"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/common"
)

const (
	NodeAgentPort           = "9001"
	multipathDir            = "/dev/mapper/"
	deviceDir               = "/dev/"
	SrcInitiatornameFile    = "/host/iscsi/initiatorname.iscsi"
	DstInitiatornameFile    = "/etc/iscsi/initiatorname.iscsi"
	DeviceWaitRetryCounts   = 5
	DeviceWaitRetryInterval = 1 * time.Second
)

func create(cli client.Client, conf *storagev1.Iscsi) error {
	if err := common.CreateNodeAnnotationsAndLabels(cli, fmt.Sprintf("%s-%s", IscsiInstanceLabelKeyPrefix, conf.Name), IscsiInstanceLabelValue, conf.Spec.Initiators); err != nil {
		return err
	}
	if err := iscsiLoginAll(cli, conf, conf.Spec.Initiators); err != nil {
		return err
	}
	if err := deployIscsiLvmd(cli, conf); err != nil {
		return err
	}
	if err := createVolumeGroup(cli, conf); err != nil {
		return err
	}
	if !checkVolumeGroup(cli, conf) {
		return errors.New("can not get volumegroup from initiators")
	}
	if err := deployIscsiCSI(cli, conf); err != nil {
		return err
	}
	if err := deployStorageClass(cli, conf); err != nil {
		return err
	}
	return nil
}

func iscsiLoginAll(cli client.Client, conf *storagev1.Iscsi, nodes []string) error {
	for _, node := range nodes {
		log.Debugf("%s: iscsi login", node)
		nodeCli, err := createNodeAgentClient(cli, node)
		if err != nil {
			return err
		}
		var username, password string
		if conf.Spec.Chap {
			username, password, err = getChap(cli, common.StorageNamespace, fmt.Sprintf("%s-%s", conf.Name, IscsiInstanceSecretSuffix))
			if err != nil {
				return err
			}
		}
		if err := replaceInitiatorname(nodeCli, SrcInitiatornameFile, DstInitiatornameFile); err != nil {
			return err
		}
		for _, target := range conf.Spec.Targets {
			if err := loginIscsi(nodeCli, target, conf.Spec.Port, conf.Spec.Iqn, username, password); err != nil {
				return err
			}
		}
		if err := reloadMultipath(nodeCli); err != nil {
			return err
		}
	}
	return nil
}

func createVolumeGroup(cli client.Client, conf *storagev1.Iscsi) error {
	node := conf.Spec.Initiators[0]
	vgName := fmt.Sprintf("%s-%s", conf.Name, VolumeGroupSuffix)
	log.Debugf("%s: create volumegroup %s", node, vgName)
	nodeCli, err := createNodeAgentClient(cli, node)
	if err != nil {
		return err
	}
	devices, err := getIscsiDevices(nodeCli, conf.Spec.Iqn)
	if err != nil {
		return fmt.Errorf("iscsi get devices failed. %v", err)
	}

	lvmdcli, err := common.CreateLvmdClientForPod(cli, node, common.StorageNamespace, fmt.Sprintf("%s-%s", conf.Name, IscsiLvmdDsSuffix))
	if err != nil {
		return fmt.Errorf("Create Lvmd client failed for host %s, %v", node, err)
	}
	defer lvmdcli.Close()
	for _, dev := range devices {
		if err := common.GenVolumeGroup(lvmdcli, dev, vgName); err != nil {
			return err
		}
	}
	return nil
}

func deployIscsiLvmd(cli client.Client, conf *storagev1.Iscsi) error {
	log.Debugf("Deploy iscsi %s lvmd", conf.Name)

	yaml, err := lvmdyaml(conf.Name)
	if err != nil {
		return err
	}
	if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	return common.WaitReady(common.DaemonSetObj(), cli, common.StorageNamespace, fmt.Sprintf("%s-%s", conf.Name, IscsiLvmdDsSuffix))
}

func checkVolumeGroup(cli client.Client, conf *storagev1.Iscsi) bool {
	log.Debugf("Check iscsi %s volumegroup ready", conf.Name)
	for i := 0; i < DeviceWaitRetryCounts; i++ {
		if err := checkReady(cli, conf); err == nil {
			return true
		}
		time.Sleep(DeviceWaitRetryInterval)
	}
	return false
}

func checkReady(cli client.Client, conf *storagev1.Iscsi) error {
	var ok int
	for _, node := range conf.Spec.Initiators {
		lvmdcli, err := common.CreateLvmdClientForPod(cli, node, common.StorageNamespace, fmt.Sprintf("%s-%s", conf.Name, IscsiLvmdDsSuffix))
		if err != nil {
			return fmt.Errorf("Create Lvmd client failed for host %s, %v", node, err)
		}
		defer lvmdcli.Close()
		vgs, err := common.GetVGs(ctx, lvmdcli)
		if err != nil {
			return fmt.Errorf("list VolumeGroup failed, %v", err)
		}
		for _, vg := range vgs.VolumeGroups {
			if vg.Name == fmt.Sprintf("%s-%s", conf.Name, VolumeGroupSuffix) {
				ok += 1
			}
		}
	}
	if ok != len(conf.Spec.Initiators) {
		return errors.New("ready node not enough")
	}
	return nil
}

func deployIscsiCSI(cli client.Client, conf *storagev1.Iscsi) error {
	log.Debugf("Deploy iscsi %s csi", conf.Name)
	yaml, err := csiyaml(conf.Name)
	if err != nil {
		return err
	}
	if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	if err := common.WaitReady(common.DeploymentObj(), cli, common.StorageNamespace, fmt.Sprintf("%s-%s", conf.Name, IscsiCSIDpSuffix)); err != nil {
		return err
	}
	return common.WaitReady(common.DaemonSetObj(), cli, common.StorageNamespace, fmt.Sprintf("%s-%s", conf.Name, IscsiCSIDsSuffix))
}

func deployStorageClass(cli client.Client, conf *storagev1.Iscsi) error {
	log.Debugf("Deploy iscsi %s storageclass", conf.Name)
	yaml, err := scyaml(conf.Name)
	if err != nil {
		return err
	}
	return helper.CreateResourceFromYaml(cli, yaml)
}
