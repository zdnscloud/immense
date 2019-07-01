package mds

import (
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
)

func Start(cli client.Client) error {
	log.Debugf("Deploy mds")
	yaml, err := mdsYaml()
	if err != nil {
		return err
	}
	return helper.CreateResourceFromYaml(cli, yaml)
}

func Stop(cli client.Client) error {
	log.Debugf("Undeploy mds")
	yaml, err := mdsYaml()
	if err != nil {
		return err
	}
	return helper.DeleteResourceFromYaml(cli, yaml)
}
