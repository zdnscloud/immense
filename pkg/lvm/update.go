package lvm

import (
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/immense/pkg/common"
)

func doDelhost(cli client.Client, cluster common.Storage) error {
	if len(cluster.Spec.Hosts) == 0 {
		return nil
	}
	log.Debugf("Delete host for storage cluster:%s, Cfg: %s", cluster.Spec.StorageType, cluster)
	if err := unInitBlocks(cli, cluster); err != nil {
		return err
	}
	return common.DeleteNodeAnnotationsAndLabels(cli, cluster)
}

func doAddhost(cli client.Client, cluster common.Storage) error {
	if len(cluster.Spec.Hosts) == 0 {
		return nil
	}
	if err := common.CreateNodeAnnotationsAndLabels(cli, cluster); err != nil {
		return err
	}
	log.Debugf("Add host for storage cluster:%s, Cfg: %s", cluster.Spec.StorageType, cluster)
	return initBlocks(cli, cluster)
}
