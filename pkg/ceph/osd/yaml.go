package osd

import (
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/common"
)

func osdYaml(fsid, host, dev, members, eps string) (string, error) {
	cfg := map[string]interface{}{
		"Namespace":     common.StorageNamespace,
		"CephInitImage": global.CephInitImage,
		"CephImage":     global.CephImage,
		"CephConfName":  global.ConfigMapName,
		"NodeName":      host,
		"OsdID":         dev,
		"FSID":          fsid,
		"Mon_Members":   members,
		"Mon_Endpoint":  eps,
	}
	return common.CompileTemplateFromMap(OsdTemp, cfg)
}
