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
	common.DeleteNodeAnnotationsAndLabels(cli, cluster)
	return nil
}

func doAddhost(cli client.Client, cluster storagev1.Cluster) error {
	if len(cluster.Spec.Hosts) == 0 {
		return nil
	}
	common.CreateNodeAnnotationsAndLabels(cli, cluster)
	waitDone(cli)
	common.WaitCSIReady(cli, common.StorageNamespace, CSIProvisionerStsName, CSIPluginDsName)
	return initBlocks(cli, cluster)
}
