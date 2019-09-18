package fscsi

import (
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/common"
)

func CSICfgYaml(id, mons string) (string, error) {
	cfg := map[string]interface{}{
		"CSIConfigmapName": global.CSIConfigmapName,
		"CephMons":         mons,
		"CephClusterID":    id,
		"StorageNamespace": common.StorageNamespace,
	}
	return common.CompileTemplateFromMap(FSconfigmapTemp, cfg)
}

func fscsiYaml() (string, error) {
	cfg := map[string]interface{}{
		"CSIPluginDsName":                 global.CSIPluginDsName,
		"CSIProvisionerStsName":           global.CSIProvisionerStsName,
		"RBACConfig":                      global.RBAC,
		"StorageCephAttacherImage":        global.CSIAttacherImage,
		"StorageCephProvisionerImage":     global.CSIProvisionerImage,
		"StorageCephDriverRegistrarImage": global.CSIDriverRegistrarImage,
		"StorageCephFsCSIImage":           global.CephFsCSIImage,
		"StorageNamespace":                common.StorageNamespace,
		"CSIConfigmapName":                global.CSIConfigmapName,
		"StorageDriverName":               global.StorageDriverName,
	}
	return common.CompileTemplateFromMap(FScsiTemp, cfg)
}

func StorageClassYaml(id, name string) (string, error) {
	cfg := map[string]interface{}{
		"CephSecretName":   	global.SecretName,
		"CephFsPool":       	global.CephFsDate,
		"CephFsName":       	global.CephFsName,
		"CephClusterID":    	id,
		"StorageNamespace": 	common.StorageNamespace,
		"StorageClassName": 	name,
		"StorageDriverName":	global.StorageDriverName,
	}
	return common.CompileTemplateFromMap(StorageClassTemp, cfg)
}
