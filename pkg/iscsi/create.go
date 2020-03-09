package iscsi

import (
	"fmt"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/common"
)

func deployIscsiInit(cli client.Client, conf *storagev1.Iscsi) error {
	log.Debugf("Deploy iscsi %s init", conf.Name)
	yaml, err := inityaml(conf.Name, conf.Spec.Target, conf.Spec.Port, conf.Spec.Iqn, conf.Spec.Chap)
	if err != nil {
		return err
	}
	if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	common.WaitDsReady(cli, common.StorageNamespace, fmt.Sprintf("%s-%s", conf.Name, IscsiInitDsNameSuffix))
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
	common.WaitDsReady(cli, common.StorageNamespace, fmt.Sprintf("%s-%s", conf.Name, IscsiLvmdDsSuffix))
	return nil
}

func checkVolumeGroup(cli client.Client, conf *storagev1.Iscsi) (bool, error) {
	log.Debugf("Check iscsi %s volumegroup ready", conf.Name)
	var ok int
	for _, node := range conf.Spec.Initiators {
		lvmdcli, err := common.CreateLvmdClientForPod(cli, node, common.StorageNamespace, fmt.Sprintf("%s-%s", conf.Name, IscsiLvmdDsSuffix))
		if err != nil {
			return false, fmt.Errorf("Create Lvmd client failed for host %s, %v", node, err)
		}
		defer lvmdcli.Close()
		vgs, err := common.GetVGs(ctx, lvmdcli)
		if err != nil {
			return false, fmt.Errorf("list VolumeGroup failed, %v", err)
		}
		for _, vg := range vgs.VolumeGroups {
			if vg.Name == fmt.Sprintf("%s-%s", conf.Name, VolumeGroupSuffix) {
				ok += 1
			}
		}
	}
	if ok == len(conf.Spec.Initiators) {
		return true, nil
	}
	return false, nil
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
	common.WaitDpReady(cli, common.StorageNamespace, fmt.Sprintf("%s-%s", conf.Name, IscsiCSIDpSuffix))
	common.WaitDsReady(cli, common.StorageNamespace, fmt.Sprintf("%s-%s", conf.Name, IscsiCSIDsSuffix))
	return nil
}

func deployStorageClass(cli client.Client, conf *storagev1.Iscsi) error {
	log.Debugf("Deploy iscsi %s storageclass", conf.Name)
	yaml, err := scyaml(conf.Name)
	if err != nil {
		return err
	}
	return helper.CreateResourceFromYaml(cli, yaml)
}
