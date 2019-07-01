package fscsi

import (
	"fmt"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/ceph/util"
	"strings"
)

func Start(cli client.Client) error {
	log.Debugf("Deploy fscsi")
	yaml, err := fscsiYaml()
	if err != nil {
		return err
	}
	if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
		return err
	}

	log.Debugf("Deploy stroageclass %s", global.StorageClassName)
	mons, err := util.GetMonIPs(cli)
	if err != nil {
		return err
	}
	monitors := strings.Replace(strings.Trim(fmt.Sprint(mons), "[]"), " ", ",", -1)
	yaml, err = storageClassYaml(monitors)
	if err != nil {
		return err
	}
	if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	return nil
}

func Stop(cli client.Client) error {
	log.Debugf("Undeploy stroageclass %s", global.StorageClassName)
	yaml, err := storageClassYaml("")
	if err != nil {
		return err
	}
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil {
		return err
	}

	log.Debugf("Undeploy fscsi")
	yaml, err = fscsiYaml()
	if err != nil {
		return err
	}
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	return nil
}
