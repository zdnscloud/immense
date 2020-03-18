package iscsi

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/common"
)

func iscsiLogoutAll(cli client.Client, conf *storagev1.Iscsi, nodes []string) error {
	for _, node := range nodes {
		log.Debugf("%s: iscsi logout and clean device mapper", node)
		nodeCli, err := createNodeAgentClient(cli, node)
		if err != nil {
			return err
		}
		devices, err := getIscsiDevices(nodeCli, conf.Spec.Iqn)
		if err != nil {
			return err
		}
		for _, dev := range devices {
			if err := cleanIscsi(nodeCli, dev); err != nil {
				return err
			}
		}
		for _, target := range conf.Spec.Targets {
			if err := logoutIscsi(nodeCli, target, conf.Spec.Port, conf.Spec.Iqn); err != nil {
				return err
			}
		}
	}
	return nil
}

func deleteVolumeGroup(cli client.Client, conf *storagev1.Iscsi) error {
	node := conf.Spec.Initiators[0]
	vgName := fmt.Sprintf("%s-%s", conf.Name, VolumeGroupSuffix)
	log.Debugf("%s: delete volumegroup %s", node, vgName)
	lvmdcli, err := common.CreateLvmdClientForPod(cli, node, common.StorageNamespace, fmt.Sprintf("%s-%s", conf.Name, IscsiLvmdDsSuffix))
	if err != nil {
		return fmt.Errorf("Create Lvmd client failed for host %s, %v", node, err)
	}
	if err := common.RemoveVG(ctx, lvmdcli, vgName); err != nil {
		return fmt.Errorf("Remove vg failed, %v", err)
	}
	return nil
}

func unDeployIscsiLvmd(cli client.Client, conf *storagev1.Iscsi) error {
	log.Debugf("Undeploy iscsi %s lvmd", conf.Name)
	yaml, err := lvmdyaml(conf.Name)
	if err != nil {
		return err
	}
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	common.WaitDsTerminated(cli, common.StorageNamespace, fmt.Sprintf("%s-%s", conf.Name, IscsiLvmdDsSuffix))
	return nil
}

func unDeployIscsiCSI(cli client.Client, conf *storagev1.Iscsi) error {
	log.Debugf("Undeploy iscsi %s csi", conf.Name)
	yaml, err := csiyaml(conf.Name)
	if err != nil {
		return err
	}
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	common.WaitDpTerminated(cli, common.StorageNamespace, fmt.Sprintf("%s-%s", conf.Name, IscsiCSIDpSuffix))
	common.WaitDsTerminated(cli, common.StorageNamespace, fmt.Sprintf("%s-%s", conf.Name, IscsiCSIDsSuffix))
	return nil
}

func unDeployStorageClass(cli client.Client, conf *storagev1.Iscsi) error {
	log.Debugf("Undeploy iscsi storageclass %s", conf.Name)
	yaml, err := scyaml(conf.Name)
	if err != nil {
		return err
	}
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	return nil
}
