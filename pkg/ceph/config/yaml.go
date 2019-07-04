package config

import (
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/common"
)

func svcYaml() (string, error) {
	cfg := map[string]interface{}{
		"Namespace": common.StorageNamespace,
		"SvcName":   global.MonSvc,
	}
	return common.CompileTemplateFromMap(MonSvcTemp, cfg)
}

func confYaml(uuid, networks, adminkey, monkey string, number int) (string, error) {
	cfg := map[string]interface{}{
		"CephConfName": global.ConfigMapName,
		"Namespace":    common.StorageNamespace,
		"MonHost":      global.MonSvc,
		"FSID":         uuid,
		"Network":      networks,
		"AdminKey":     adminkey,
		"MonKey":       monkey,
		"Replication":  number,
	}
	return common.CompileTemplateFromMap(ConfigMapTemp, cfg)
}

func secretYaml(user, secret string) (string, error) {
	cfg := map[string]interface{}{
		"CephSecretName":      global.SecretName,
		"CephAdminUserEncode": user,
		"CephAdminKeyEncode":  secret,
		"StorageNamespace":    common.StorageNamespace,
	}
	return common.CompileTemplateFromMap(SecretTemp, cfg)
}
