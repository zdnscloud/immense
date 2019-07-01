package mds

import (
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/common"
)

func mdsYaml() (string, error) {
	cfg := map[string]interface{}{
		"Namespace":            common.StorageNamespace,
		"CephInitImage":        global.CephInitImage,
		"CephImage":            global.CephImage,
		"CEPHFS_NAME":          global.CephFsName,
		"CEPHFS_DATA_POOL":     global.CephFsDate,
		"CEPHFS_METADATA_POOL": global.CephFsMetadata,
		"MdsNum":               global.MdsNum,
	}
	return common.CompileTemplateFromMap(MdsTemp, cfg)
}
