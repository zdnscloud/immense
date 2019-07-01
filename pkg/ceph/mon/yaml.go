package mon

import (
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/common"
)

func monYaml(networks string) (string, error) {
	cfg := map[string]interface{}{
		"Namespace":     common.StorageNamespace,
		"Network":       networks,
		"MonNum":        global.MonNum,
		"CephInitImage": global.CephInitImage,
		"CephImage":     global.CephImage,
	}
	return common.CompileTemplateFromMap(MonTemp, cfg)
}
