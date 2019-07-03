package fscsi

import (
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/common"
)

func fscsiYaml() (string, error) {
	cfg := map[string]interface{}{
		"RBACConfig":                      global.RBAC,
		"StorageCephAttacherImage":        global.CSIAttacherImage,
		"StorageCephProvisionerImage":     global.CSIProvisionerImage,
		"StorageCephDriverRegistrarImage": global.CSIDriverRegistrarImage,
		"StorageCephFsCSIImage":           global.CephFsCSIImage,
		"StorageNamespace":                common.StorageNamespace,
	}
	return common.CompileTemplateFromMap(FScsiTemp, cfg)
}

func StorageClassYaml(monitors string) (string, error) {
	cfg := map[string]interface{}{
		"CephClusterMonitors": monitors,
		"CephSecretName":      global.SecretName,
		"CephFsPool":          global.CephFsDate,
		"StorageNamespace":    common.StorageNamespace,
		"StorageClassName":    global.StorageClassName,
	}
	return common.CompileTemplateFromMap(StorageClassTemp, cfg)
}
