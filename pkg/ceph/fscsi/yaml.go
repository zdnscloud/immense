package fscsi

import (
	"fmt"

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

func fscsiYaml(name string) (string, error) {
	cfg := map[string]interface{}{
		"RBACConfig":                      common.RBACConfig,
		"StorageNamespace":                common.StorageNamespace,
		"StorageCephAttacherImage":        common.CSIAttacherImage,
		"StorageCephResizerImage":         common.CSIResizerImage,
		"StorageCephProvisionerImage":     common.CSIProvisionerImage,
		"StorageCephDriverRegistrarImage": common.CSIDriverRegistrarImage,
		"CSIPluginDsName":                 global.CSIPluginDsName,
		"CSIProvisionerStsName":           global.CSIProvisionerStsName,
		"StorageCephFsCSIImage":           global.CephFsCSIImage,
		"CSIConfigmapName":                global.CSIConfigmapName,
		"StorageDriverName":               fmt.Sprintf("%s.%s", name, global.CephFsDriverSuffix),
	}
	return common.CompileTemplateFromMap(FScsiTemp, cfg)
}

func StorageClassYaml(id, name string) (string, error) {
	cfg := map[string]interface{}{
		"StorageNamespace":  common.StorageNamespace,
		"StorageClassName":  name,
		"StorageDriverName": fmt.Sprintf("%s.%s", name, global.CephFsDriverSuffix),
		"CephSecretName":    global.SecretName,
		"CephFsPool":        global.CephFsDate,
		"CephFsName":        global.CephFsName,
		"CephClusterID":     id,
	}
	return common.CompileTemplateFromMap(StorageClassTemp, cfg)
}
