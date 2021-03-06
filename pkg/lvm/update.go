package lvm

import (
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/common"
)

func doDelhost(cli client.Client, cluster storagev1.Cluster) error {
	if len(cluster.Spec.Hosts) == 0 {
		return nil
	}
	if err := unInitBlocks(cli, cluster); err != nil {
		return err
	}
	if err := common.DeleteNodeAnnotationsAndLabels(cli, common.StorageHostLabels, cluster.Spec.StorageType, cluster.Spec.Hosts); err != nil {
		return err
	}
	return nil
}

func doAddhost(cli client.Client, cluster storagev1.Cluster) error {
	if len(cluster.Spec.Hosts) == 0 {
		return nil
	}
	if err := common.CreateNodeAnnotationsAndLabels(cli, common.StorageHostLabels, cluster.Spec.StorageType, cluster.Spec.Hosts); err != nil {
		return err
	}
	if err := common.WaitReady(common.DaemonSetObj(), cli, common.StorageNamespace, LvmdDsName); err != nil {
		return err
	}
	if err := common.WaitReady(common.DaemonSetObj(), cli, common.StorageNamespace, CSIPluginDsName); err != nil {
		return err
	}
	return initBlocks(cli, cluster)
}
