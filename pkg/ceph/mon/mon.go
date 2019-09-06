package mon

import (
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/ceph/util"
)

func Start(cli client.Client, networks string) error {
	log.Debugf("Deploy mon")
	yaml, err := monYaml(networks)
	if err != nil {
		return err
	}
	if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	util.WaitDpReady(cli, global.MonDpName)
	return nil
}

func Stop(cli client.Client, networks string) error {
	log.Debugf("Undeploy mon")
	yaml, err := monYaml(networks)
	if err != nil {
		return err
	}
	return helper.DeleteResourceFromYaml(cli, yaml)
}
