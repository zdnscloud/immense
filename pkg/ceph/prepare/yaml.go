package prepare

import (
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/common"
)

func prepareYaml(node, devs string) (string, error) {
	cfg := map[string]interface{}{
		"Namespace":     common.StorageNamespace,
		"CephInitImage": global.CephInitImage,
		"CephImage":     global.CephImage,
		"CephConfName":  global.ConfigMapName,
		"NodeName":      node,
		"Devices":       devs,
	}
	return common.CompileTemplateFromMap(PrepareTemp, cfg)
}
