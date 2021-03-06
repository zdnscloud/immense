package mgr

import (
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/common"
)

func mgrYaml() (string, error) {
	cfg := map[string]interface{}{
		"Namespace":     common.StorageNamespace,
		"CephInitImage": global.CephInitImage,
		"CephImage":     global.CephImage,
		"CephConfName":  global.ConfigMapName,
		"MgrNum":        global.MgrNum,
		"MgrDpName":     global.MgrDpName,
	}
	return common.CompileTemplateFromMap(MgrTemp, cfg)
}
