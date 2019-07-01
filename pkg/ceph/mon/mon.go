package mon

import (
	"errors"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/ceph/util"
	"time"
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
	ready, err := check(cli)
	if err != nil {
		return err
	}
	if !ready {
		return errors.New("Timeout. Ceph cluster has not ready")
	}
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

func check(cli client.Client) (bool, error) {
	log.Debugf("Wait all mon running")
	for i := 0; i < 60; i++ {
		mons, err := util.GetMonIPs(cli)
		if err != nil {
			return false, err
		}
		if len(mons) == global.MonNum {
			return true, nil
		}
		time.Sleep(5 * time.Second)
	}
	return false, errors.New("Timeout. Mon has not ready")
}
