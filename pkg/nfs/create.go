package nfs

import (
	"fmt"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/common"
)

func deployNfsCSI(cli client.Client, conf *storagev1.Nfs) error {
	log.Debugf("Deploy nfs %s csi", conf.Name)

	yaml, err := csiyaml(conf.Name, conf.Spec.Server, conf.Spec.Path)
	if err != nil {
		return err
	}
	if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	common.WaitDpReady(cli, common.StorageNamespace, fmt.Sprintf("%s-%s", conf.Name, NfsCSIDpSuffix))
	return nil
}

func deployStorageClass(cli client.Client, conf *storagev1.Nfs) error {
	log.Debugf("Deploy nfs %s storageclass", conf.Name)
	yaml, err := scyaml(conf.Name)
	if err != nil {
		return err
	}
	return helper.CreateResourceFromYaml(cli, yaml)
}
