package mds

import (
	"fmt"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/ceph/util"
	"strings"
)

func Start(cli client.Client, fsid string, monsvc map[string]string, size, pgnum int) error {
	monHosts := util.GetMonHosts(monsvc)
	monMembers := strings.Replace(strings.Trim(fmt.Sprint(global.MonMembers), "[]"), " ", ",", -1)
	log.Debugf("Deploy mds")
	yaml, err := mdsYaml(fsid, monHosts, monMembers, size, pgnum)
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
	var size, pgnum int
	var fsid, monHosts, monMembers string
	yaml, err := mdsYaml(fsid, monHosts, monMembers, size, pgnum)
	if err != nil {
		return err
	}
	return helper.DeleteResourceFromYaml(cli, yaml)
}
