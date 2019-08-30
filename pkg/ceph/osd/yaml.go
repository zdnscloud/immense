package osd

import (
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/common"
)

func osdYaml(host, dev string) (string, error) {
	cfg := map[string]interface{}{
		"Namespace":     common.StorageNamespace,
		"CephInitImage": global.CephInitImage,
		"CephImage":     global.CephImage,
		"CephConfName":  global.ConfigMapName,
		"NodeName":      host,
		"OsdID":         dev,
	}
	return common.CompileTemplateFromMap(OsdTemp, cfg)
}
