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
		"IscsiInitImage":          IscsiInitImage,
		"IscsiLvmdImage":          IscsiLvmdImage,
		"IscsiCSIDsName":          fmt.Sprintf("%s-%s", name, IscsiCSIDsSuffix),
		"IscsiCSIDpName":          fmt.Sprintf("%s-%s", name, IscsiCSIDpSuffix),
		"VolumeGroup":             fmt.Sprintf("%s-%s", name, VolumeGroupSuffix),
		"IscsiDriverName":         fmt.Sprintf("%s.%s", name, IscsiDriverSuffix),
		"LvmdDsName":              fmt.Sprintf("%s-%s", name, IscsiLvmdDsSuffix),
	}
	return common.CompileTemplateFromMap(IscsiCSITemplate, cfg)
}

func inityaml(name, host, port, iqn string, chap bool) (string, error) {
	cfg := map[string]interface{}{
		"CHAPConfig":              chap,
		"Instance":                name,
		"IscsiInitDsName":         fmt.Sprintf("%s-%s", name, IscsiInitDsNameSuffix),
		"IscsiInstanceSecret":     fmt.Sprintf("%s-%s", name, IscsiInstanceSecretSuffix),
		"IscsiInstanceLabelKey":   fmt.Sprintf("%s-%s", IscsiInstanceLabelKeyPrefix, name),
		"IscsiInstanceLabelValue": IscsiInstanceLabelValue,
		"StorageNamespace":        common.StorageNamespace,
		"IscsiInitImage":          IscsiInitImage,
		"TargetHost":              host,
		"TargetPort":              port,
		"TargetIqn":               iqn,
		"VolumeGroup":             fmt.Sprintf("%s-%s", name, VolumeGroupSuffix),
	}
	return common.CompileTemplateFromMap(IscsiInitTemplate, cfg)
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

func jobyaml(host, iqn string) (string, error) {
	cfg := map[string]interface{}{
		"StorageNamespace": common.StorageNamespace,
		"IscsiStopJobName": fmt.Sprintf("%s-%s", host, IscsiStopJobSuffix),
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
