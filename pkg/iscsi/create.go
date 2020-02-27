package iscsi

import (
	"fmt"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/common"
)

func deployIscsiLvmd(cli client.Client, conf *storagev1.Iscsi) error {
	log.Debugf("Deploy iscsi %s lvmd", conf.Name)

	yaml, err := lvmdyaml(conf.Name, conf.Spec.Target, conf.Spec.Port, conf.Spec.Iqn)
	if err != nil {
		return err
	}
	if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	common.WaitDsReady(cli, common.StorageNamespace, fmt.Sprintf("%s-%s", conf.Name, IscsiLvmdDsSuffix))
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
	common.WaitStsReady(cli, common.StorageNamespace, fmt.Sprintf("%s-%s", conf.Name, IscsiCSIStsSuffix))
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
