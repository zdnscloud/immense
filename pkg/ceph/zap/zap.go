package zap

import (
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
	check(cli, host, dev)
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

func check(cli client.Client, host, dev string) {
	log.Debugf("Wait zap done %s:%s", host, dev)
	name := "ceph-job-zap-" + host + "-" + dev
	var ready bool
	for !ready {
		time.Sleep(10 * time.Second)
		ok, err := util.CheckPodPhase(cli, name, "Succeeded")
		if err != nil || !ok {
			continue
		}
		ready = true
	}
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
