package osd

import (
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/common"
)

func osdYaml(fsid, host, dev string) (string, error) {
	cfg := map[string]interface{}{
		"Namespace":     common.StorageNamespace,
		"CephInitImage": global.CephInitImage,
		"CephImage":     global.CephImage,
		"CephConfName":  global.ConfigMapName,
		"NodeName":      host,
		"OsdID":         dev,
		"FSID":          fsid,
	}
	return common.CompileTemplateFromMap(OsdTemp, cfg)
}
