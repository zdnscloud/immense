package iscsi

import (
	"fmt"

	"github.com/zdnscloud/immense/pkg/common"
)

func csiyaml(name string) (string, error) {
	cfg := map[string]interface{}{
		"RBACConfig":              common.RBACConfig,
		"LabelKey":                common.StorageHostLabels,
		"LabelValue":              fmt.Sprintf("%s-%s", StorageType, name),
		"StorageNamespace":        common.StorageNamespace,
		"CSIAttacherImage":        common.CSIAttacherImage,
		"CSIResizerImage":         common.CSIResizerImage,
		"CSIProvisionerImage":     common.CSIProvisionerImage,
		"CSIDriverRegistrarImage": common.CSIDriverRegistrarImage,
		"IscsiPluginImage":        IscsiPluginImage,
		"IscsiInitImage":          IscsiInitImage,
		"IscsiLvmdImage":          IscsiLvmdImage,
		"IscsiCSIDsName":          fmt.Sprintf("%s-%s", name, IscsiCSIDsSuffix),
		"IscsiCSIStsName":         fmt.Sprintf("%s-%s", name, IscsiCSIStsSuffix),
		"VolumeGroup":             fmt.Sprintf("%s-%s", name, VolumeGroupSuffix),
		"IscsiDriverName":         fmt.Sprintf("%s.%s", name, IscsiDriverSuffix),
	}
	return common.CompileTemplateFromMap(IscsiCSITemplate, cfg)
}

func lvmdyaml(name, host, port, iqn string) (string, error) {
	cfg := map[string]interface{}{
		"RBACConfig":       common.RBACConfig,
		"LabelKey":         common.StorageHostLabels,
		"LabelValue":       fmt.Sprintf("%s-%s", StorageType, name),
		"StorageNamespace": common.StorageNamespace,
		"IscsiLvmdDsName":  fmt.Sprintf("%s-%s", name, IscsiLvmdDsSuffix),
		"IscsiLvmdImage":   IscsiLvmdImage,
		"IscsiInitImage":   IscsiInitImage,
		"TargetHost":       host,
		"TargetPort":       port,
		"TargetIqn":        iqn,
		"VolumeGroup":      fmt.Sprintf("%s-%s", name, VolumeGroupSuffix),
	}
	return common.CompileTemplateFromMap(IscsiLvmdTemplate, cfg)
}

func jobyaml(host, iqn string) (string, error) {
	cfg := map[string]interface{}{
		"StorageNamespace": common.StorageNamespace,
		"IscsiStopJobName": fmt.Sprintf("%s-%s", IscsiStopJobPrefix, host),
		"IscsiInitImage":   IscsiInitImage,
		"TargetIqn":        iqn,
		"Host":             host,
	}
	return common.CompileTemplateFromMap(IscsiStopJobTemplate, cfg)
}

func scyaml(name string) (string, error) {
	cfg := map[string]interface{}{
		"StorageClassName": name,
		"IscsiDriverName":  fmt.Sprintf("%s.%s", name, IscsiDriverSuffix),
	}
	return common.CompileTemplateFromMap(StorageClassTemp, cfg)
}
