package mgr

import (
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/common"
)

func Start(cli client.Client) error {
	log.Debugf("Deploy mgr")
	yaml, err := mgrYaml()
	if err != nil {
		return err
	}
	if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	return common.WaitDpReady(cli, common.StorageNamespace, global.MgrDpName)
}

func Stop(cli client.Client) error {
	log.Debugf("Undeploy mgr")
	yaml, err := mgrYaml()
	if err != nil {
		return err
	}
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	return common.WaitDpTerminated(cli, common.StorageNamespace, global.MgrDpName)
}
