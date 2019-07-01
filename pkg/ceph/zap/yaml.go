package zap

import (
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/common"
)

func osdZapYaml(node, dev string) (string, error) {
	cfg := map[string]interface{}{
		"Namespace": common.StorageNamespace,
		"CephImage": global.CephImage,
		"NodeName":  node,
		"OsdID":     dev,
	}
	return common.CompileTemplateFromMap(OsdZapTemp, cfg)
}
