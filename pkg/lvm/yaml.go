package lvm

import (
	"github.com/zdnscloud/immense/pkg/common"
)

func csiyaml(name string) (string, error) {
	cfg := map[string]interface{}{
		"CSIProvisionerStsName":          CSIProvisionerStsName,
		"CSIPluginDsName":                CSIPluginDsName,
		"RBACConfig":                     common.RBACConfig,
		"LabelKey":                       common.StorageHostLabels,
		"LabelValue":                     common.LvmLabelsValue,
		"StorageNamespace":               common.StorageNamespace,
		"StorageLvmAttacherImage":        "quay.io/k8scsi/csi-attacher:v1.0.1",
		"StorageLvmProvisionerImage":     "quay.io/k8scsi/csi-provisioner:v1.0.1",
		"StorageLvmDriverRegistrarImage": "quay.io/k8scsi/csi-node-driver-registrar:v1.0.2",
		"StorageLvmCSIImage":             "zdnscloud/lvmcsi:v0.6.2",
		"StorageClassName":               name,
		"StorageDriverName":              StorageDriverName,
	}
	return common.CompileTemplateFromMap(LvmCSITemplate, cfg)
}

func lvmdyaml() (string, error) {
	cfg := map[string]interface{}{
		"LvmdDsName":       LvmdDsName,
		"RBACConfig":       common.RBACConfig,
		"LabelValue":       common.LvmLabelsValue,
		"LabelKey":         common.StorageHostLabels,
		"StorageLvmdImage": "zdnscloud/lvmd:v0.5",
		"StorageNamespace": common.StorageNamespace,
	}
	return common.CompileTemplateFromMap(LvmdTemplate, cfg)
}

func scyaml(name string) (string, error) {
	cfg := map[string]interface{}{
		"StorageClassName":  name,
		"StorageDriverName": StorageDriverName,
	}
	return common.CompileTemplateFromMap(StorageClassTemp, cfg)
}
