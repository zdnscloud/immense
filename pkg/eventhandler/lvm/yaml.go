package lvm

import (
	"github.com/zdnscloud/immense/pkg/eventhandler/common"
)

func csiyaml() (string, error) {
	cfg := map[string]interface{}{
		"AoDNamespace":                   "yes",
		"RBACConfig":                     common.RBACConfig,
		"LabelKey":                       common.StorageHostLabels,
		"LabelValue":                     NodeLabelValue,
		"StorageNamespace":               common.StorageNamespace,
		"StorageLvmAttacherImage":        "quay.io/k8scsi/csi-attacher:v1.0.0",
		"StorageLvmProvisionerImage":     "quay.io/k8scsi/csi-provisioner:v1.0.0",
		"StorageLvmDriverRegistrarImage": "quay.io/k8scsi/csi-node-driver-registrar:v1.0.2",
		"StorageLvmCSIImage":             "zdnscloud/lvmcsi:v0.5",
		"StorageClassName":               "lvm",
	}
	return common.CompileTemplateFromMap(LvmCSITemplate, cfg)
}

func lvmdyaml() (string, error) {
	cfg := map[string]interface{}{
		"AoDNamespace":     "yes",
		"RBACConfig":       common.RBACConfig,
		"LabelKey":         common.StorageHostLabels,
		"LabelValue":       NodeLabelValue,
		"StorageNamespace": common.StorageNamespace,
		"StorageLvmdImage": "zdnscloud/lvmd:v0.96",
	}
	return common.CompileTemplateFromMap(LvmDTemplate, cfg)
}
