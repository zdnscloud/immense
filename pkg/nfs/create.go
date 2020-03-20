package nfs

import (
	"fmt"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/common"
)

func create(cli client.Client, conf *storagev1.Nfs) error {
	if err := deployNfsCSI(cli, conf); err != nil {
		return err
	}
	if err := deployStorageClass(cli, conf); err != nil {
		return err
	}
	if err := AddFinalizer(cli, conf.Name, common.StoragePrestopHookFinalizer); err != nil {
		return err
	}
	return nil
}

func deployNfsCSI(cli client.Client, conf *storagev1.Nfs) error {
	log.Debugf("Deploy nfs %s csi", conf.Name)

	yaml, err := csiyaml(conf.Name, conf.Spec.Server, conf.Spec.Path)
	if err != nil {
		return err
	}
	if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	return common.WaitDpReady(cli, common.StorageNamespace, fmt.Sprintf("%s-%s", conf.Name, NfsCSIDpSuffix))
}

func deployStorageClass(cli client.Client, conf *storagev1.Nfs) error {
	log.Debugf("Deploy nfs %s storageclass", conf.Name)
	yaml, err := scyaml(conf.Name)
	if err != nil {
		return err
	}
	return helper.CreateResourceFromYaml(cli, yaml)
}
