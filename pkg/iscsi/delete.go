package iscsi

import (
	"fmt"

	"github.com/zdnscloud/cement/errgroup"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/common"
)

func unDeployIscsiInit(cli client.Client, conf *storagev1.Iscsi) error {
	log.Debugf("Undeploy iscsi %s init", conf.Name)
	yaml, err := inityaml(conf.Name, conf.Spec.Target, conf.Spec.Port, conf.Spec.Iqn, conf.Spec.Chap)
	if err != nil {
		return err
	}
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	common.WaitDsTerminated(cli, common.StorageNamespace, fmt.Sprintf("%s-%s", conf.Name, IscsiInitDsNameSuffix))
	return nil
}

func unDeployIscsiLvmd(cli client.Client, conf *storagev1.Iscsi) error {
	log.Debugf("Undeploy iscsi %s lvmd", conf.Name)
	yaml, err := lvmdyaml(conf.Name)
	if err != nil {
		return err
	}
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	common.WaitDsTerminated(cli, common.StorageNamespace, fmt.Sprintf("%s-%s", conf.Name, IscsiLvmdDsSuffix))
	return nil
}

func unDeployIscsiCSI(cli client.Client, conf *storagev1.Iscsi) error {
	log.Debugf("Undeploy iscsi %s csi", conf.Name)
	yaml, err := csiyaml(conf.Name)
	if err != nil {
		return err
	}
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	common.WaitStsTerminated(cli, common.StorageNamespace, fmt.Sprintf("%s-%s", conf.Name, IscsiCSIStsSuffix))
	common.WaitDsTerminated(cli, common.StorageNamespace, fmt.Sprintf("%s-%s", conf.Name, IscsiCSIDsSuffix))
	return nil
}

func unDeployStorageClass(cli client.Client, conf *storagev1.Iscsi) error {
	log.Debugf("Undeploy iscsi storageclass %s", conf.Name)
	yaml, err := scyaml(conf.Name)
	if err != nil {
		return err
	}
	return helper.DeleteResourceFromYaml(cli, yaml)
}

func logoutIscsi(cli client.Client, conf *storagev1.Iscsi) error {
	_, err := errgroup.Batch(
		conf.Spec.Initiators,
		func(o interface{}) (interface{}, error) {
			host := o.(string)
			return nil, jobDo(cli, host, conf.Spec.Iqn)
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func jobDo(cli client.Client, host, iqn string) error {
	log.Debugf("Logout iscsi on host %s", host)
	yaml, err := jobyaml(host, iqn)
	if err != nil {
		return err
	}
	if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	common.WaitPodSucceeded(cli, common.StorageNamespace, fmt.Sprintf("%s-%s", host, IscsiStopJobSuffix))
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	return nil
}
