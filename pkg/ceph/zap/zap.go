package zap

import (
	"errors"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	"github.com/zdnscloud/immense/pkg/ceph/util"
	"time"
)

func Do(cli client.Client, host, dev string) error {
	log.Debugf("Init device %s:%s", host, dev)
	if err := create(cli, host, dev); err != nil {
		return err
	}
	ready, err := check(cli, host, dev)
	if err != nil {
		return err
	}
	if !ready {
		return errors.New("Zap device failed!")
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

func check(cli client.Client, host, dev string) (bool, error) {
	name := "ceph-job-zap-" + host + "-" + dev
	for i := 0; i < 60; i++ {
		suc, err := util.IsPodSucceeded(cli, name)
		if err != nil {
			continue
		}
		if suc {
			return true, nil
		}
		time.Sleep(5 * time.Second)
	}
	return false, nil
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
