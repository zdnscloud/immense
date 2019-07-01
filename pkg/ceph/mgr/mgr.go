package mgr

import (
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
)

func Start(cli client.Client) error {
	log.Debugf("Deploy mgr")
	yaml, err := mgrYaml()
	if err != nil {
		return err
	}
	return helper.CreateResourceFromYaml(cli, yaml)
}

func Stop(cli client.Client) error {
	log.Debugf("Undeploy mgr")
	yaml, err := mgrYaml()
	if err != nil {
		return err
	}
	return helper.DeleteResourceFromYaml(cli, yaml)
}
