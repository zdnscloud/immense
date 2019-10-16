package mds

import (
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/common"
)

func mdsYaml(pgnum int) (string, error) {
	cfg := map[string]interface{}{
		"Namespace":            common.StorageNamespace,
		"CephInitImage":        global.CephInitImage,
		"CephImage":            global.CephImage,
		"CephConfName":         global.ConfigMapName,
		"CEPHFS_NAME":          global.CephFsName,
		"CEPHFS_DATA_POOL":     global.CephFsDate,
		"CEPHFS_METADATA_POOL": global.CephFsMetadata,
		"MdsNum":               global.MdsNum,
		"MdsDpName":            global.MdsDpName,
		"PgNum":                pgnum,
	}
	return common.CompileTemplateFromMap(MdsTemp, cfg)
}
