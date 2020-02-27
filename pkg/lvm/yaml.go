package lvm

import (
	"github.com/zdnscloud/immense/pkg/common"
)

func csiyaml(name string) (string, error) {
	cfg := map[string]interface{}{
		"StorageNamespace":               common.StorageNamespace,
		"RBACConfig":                     common.RBACConfig,
		"LabelKey":                       common.StorageHostLabels,
		"LabelValue":                     StorageType,
		"StorageLvmAttacherImage":        common.CSIAttacherImage,
		"StorageLvmResizerImage":         common.CSIResizerImage,
		"StorageLvmProvisionerImage":     common.CSIProvisionerImage,
		"StorageLvmDriverRegistrarImage": common.CSIDriverRegistrarImage,
		"StorageLvmCSIImage":             StorageLvmCSIImage,
		"CSIProvisionerStsName":          CSIProvisionerStsName,
		"CSIPluginDsName":                CSIPluginDsName,
		"StorageClassName":               name,
		"StorageDriverName":              StorageDriverName,
		"VolumeGroup":                    VolumeGroup,
	}
	return common.CompileTemplateFromMap(LvmCSITemplate, cfg)
}

func lvmdyaml() (string, error) {
	cfg := map[string]interface{}{
		"StorageNamespace": common.StorageNamespace,
		"RBACConfig":       common.RBACConfig,
		"LabelKey":         common.StorageHostLabels,
		"LabelValue":       StorageType,
		"LvmdDsName":       LvmdDsName,
		"StorageLvmdImage": StorageLvmdImage,
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
