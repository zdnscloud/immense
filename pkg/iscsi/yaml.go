package iscsi

import (
	"fmt"

	"github.com/zdnscloud/immense/pkg/common"
)

func csiyaml(name string) (string, error) {
	cfg := map[string]interface{}{
		"RBACConfig":              common.RBACConfig,
		"IscsiInstanceLabelKey":   fmt.Sprintf("%s-%s", IscsiInstanceLabelKeyPrefix, name),
		"IscsiInstanceLabelValue": IscsiInstanceLabelValue,
		"StorageNamespace":        common.StorageNamespace,
		"CSIAttacherImage":        common.CSIAttacherImage,
		"CSIResizerImage":         common.CSIResizerImage,
		"CSIProvisionerImage":     common.CSIProvisionerImage,
		"CSIDriverRegistrarImage": common.CSIDriverRegistrarImage,
		"IscsiPluginImage":        IscsiPluginImage,
		"Instance":                name,
		"IscsiLvmdImage":          IscsiLvmdImage,
		"IscsiCSIDsName":          fmt.Sprintf("%s-%s", name, IscsiCSIDsSuffix),
		"IscsiCSIDpName":          fmt.Sprintf("%s-%s", name, IscsiCSIDpSuffix),
		"VolumeGroup":             fmt.Sprintf("%s-%s", name, VolumeGroupSuffix),
		"IscsiDriverName":         fmt.Sprintf("%s.%s", name, IscsiDriverSuffix),
		"LvmdDsName":              fmt.Sprintf("%s-%s", name, IscsiLvmdDsSuffix),
	}
	return common.CompileTemplateFromMap(IscsiCSITemplate, cfg)
}

func lvmdyaml(name string) (string, error) {
	cfg := map[string]interface{}{
		"RBACConfig":              common.RBACConfig,
		"Instance":                name,
		"IscsiInstanceLabelKey":   fmt.Sprintf("%s-%s", IscsiInstanceLabelKeyPrefix, name),
		"IscsiInstanceLabelValue": IscsiInstanceLabelValue,
		"StorageNamespace":        common.StorageNamespace,
		"IscsiLvmdDsName":         fmt.Sprintf("%s-%s", name, IscsiLvmdDsSuffix),
		"IscsiLvmdImage":          IscsiLvmdImage,
	}
	return common.CompileTemplateFromMap(IscsiLvmdTemplate, cfg)
}

func scyaml(name string) (string, error) {
	cfg := map[string]interface{}{
		"StorageClassName": name,
		"IscsiDriverName":  fmt.Sprintf("%s.%s", name, IscsiDriverSuffix),
	}
	return common.CompileTemplateFromMap(StorageClassTemp, cfg)
}
