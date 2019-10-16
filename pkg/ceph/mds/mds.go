package mds

import (
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/ceph/util"
)

func Start(cli client.Client, pgnum int) error {
	log.Debugf("Deploy mds")
	yaml, err := mdsYaml(pgnum)
	if err != nil {
		return err
	}
	if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	util.WaitDpReady(cli, global.MdsDpName)
	return nil
}

func Stop(cli client.Client) error {
	log.Debugf("Undeploy mds")
	var pgnum int
	yaml, err := mdsYaml(pgnum)
	if err != nil {
		return err
	}
	return helper.DeleteResourceFromYaml(cli, yaml)
}
