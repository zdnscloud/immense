package zap

import (
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	"github.com/zdnscloud/immense/pkg/common"
)

func Do(cli client.Client, host, dev string) error {
	log.Debugf("Init device %s:%s", host, dev)
	if err := create(cli, host, dev); err != nil {
		return err
	}
	name := "ceph-job-zap-" + host + "-" + dev
	if err := common.WaitPodSucceeded(cli, common.StorageNamespace, name); err != nil {
		return err
	}
	if err := delete(cli, host, dev); err != nil {
		return err
	}
	return nil
}

func create(cli client.Client, host, dev string) error {
	yaml, err := osdZapYaml(host, dev)
	if err != nil {
		return err
	}
	if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	return nil
}

func delete(cli client.Client, host, dev string) error {
	yaml, err := osdZapYaml(host, dev)
	if err != nil {
		return err
	}
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	return nil
}
