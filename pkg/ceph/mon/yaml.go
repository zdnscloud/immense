package mon

import (
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/common"
)

func monYaml(fsid, id, addr, monHosts, monMembers string) (string, error) {
	cfg := map[string]interface{}{
		"ID":                 id,
		"FSID":               fsid,
		"MON_HOSTS":          monHosts,
		"MON_MEMBERS":        monMembers,
		"MonSvcAddr":         addr,
		"MonSvc":             global.MonSvc,
		"Namespace":          common.StorageNamespace,
		"ServiceAccountName": global.ServiceAccountName,
		"CephConfName":       global.ConfigMapName,
		"CephInitImage":      global.CephInitImage,
		"CephImage":          global.CephImage,
		"MonPortV1":          global.MonPortV1,
		"MonPortV2":          global.MonPortV2,
	}
	return common.CompileTemplateFromMap(MonTemp, cfg)
}
