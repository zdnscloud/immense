package config

import (
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/common"
)

func svcYaml(id string) (string, error) {
	cfg := map[string]interface{}{
		"Namespace": common.StorageNamespace,
		"MonSvc":    global.MonSvc,
		"MonPortV1": global.MonPortV1,
		"MonPortV2": global.MonPortV2,
		"MonID":     id,
	}
	return common.CompileTemplateFromMap(MonSvcTemp, cfg)
}

func confYaml(uuid, adminkey, monkey, endpoint, members string, number int) (string, error) {
	cfg := map[string]interface{}{
		"CephConfName": global.ConfigMapName,
		"Namespace":    common.StorageNamespace,
		"MonHost":      global.MonSvc,
		"FSID":         uuid,
		"AdminKey":     adminkey,
		"MonKey":       monkey,
		"Replication":  number,
		"MonEp":        endpoint,
		"MonMembers":   members,
	}
	return common.CompileTemplateFromMap(ConfigMapTemp, cfg)
}

func secretYaml(user, secret string) (string, error) {
	cfg := map[string]interface{}{
		"CephSecretName":   global.SecretName,
		"CephAdminUser":    user,
		"CephAdminKey":     secret,
		"StorageNamespace": common.StorageNamespace,
	}
	return common.CompileTemplateFromMap(SecretTemp, cfg)
}

func saYaml() (string, error) {
	cfg := map[string]interface{}{
		"RBACConfig":       common.RBACConfig,
		"CephSAName":       global.ServiceAccountName,
		"StorageNamespace": common.StorageNamespace,
	}
	return common.CompileTemplateFromMap(ServiceAccountTemp, cfg)
}
