package lvm

import (
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/eventhandler/common"
)

func Delete(cli client.Client, cluster *storagev1.Cluster) error {
	if err := undeploy(cli, cluster); err != nil {
		return err
	}
	return common.DeleteNodeAnnotationsAndLabels(cli, cluster, NodeLabelValue)
}

func undeploy(cli client.Client, cluster *storagev1.Cluster) error {
	log.Debugf("Undeploy for storage cluster:%s", cluster.Spec.StorageType)
	cfg := map[string]interface{}{
		"AoDNamespace":                   "no",
		"RBACConfig":                     common.RBACConfig,
		"LabelKey":                       common.StorageHostLabels,
		"LabelValue":                     NodeLabelValue,
		"StorageNamespace":               common.StorageNamespace,
		"StorageLvmdImage":               "zdnscloud/lvmd:v0.4",
		"StorageLvmAttacherImage":        "quay.io/k8scsi/csi-attacher:v1.0.0",
		"StorageLvmProvisionerImage":     "quay.io/k8scsi/csi-provisioner:v1.0.0",
		"StorageLvmDriverRegistrarImage": "quay.io/k8scsi/csi-node-driver-registrar:v1.0.2",
		"StorageLvmCSIImage":             "zdnscloud/lvmcsi:v0.5",
		"StorageClassName":               "lvm",
	}
	yaml, err := common.CompileTemplateFromMap(LvmDTemplate, cfg)
	if err != nil {
		return err
	}
	return helper.DeleteResourceFromYaml(cli, yaml)
}
