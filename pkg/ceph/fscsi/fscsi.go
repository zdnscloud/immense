package fscsi

import (
	"strings"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/ceph/util"
	"github.com/zdnscloud/immense/pkg/common"
)

func Start(cli client.Client, id, name string, monsvc map[string]string) error {
	var mons string
	for _, ip := range monsvc {
		mon := "\"" + ip + ":" + global.MonPortV1 + "\","
		mons += mon
	}
	ms := strings.TrimRight(mons, ",")

	exist, err := util.CheckConfigMap(cli, common.StorageNamespace, global.CSIConfigmapName)
	if !exist || err != nil {
		log.Debugf("Deploy csi-cfg")
		yaml, err := CSICfgYaml(id, ms)
		if err != nil {
			return err
		}
		if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
			return err
		}
	}
	log.Debugf("Deploy fscsi")
	yaml, err := fscsiYaml()
	if err != nil {
		return err
	}
	if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	common.WaitCSIReady(cli, common.StorageNamespace, global.CSIProvisionerStsName, global.CSIPluginDsName)

	exist, err = util.CheckStorageclass(cli, name)
	if !exist || err != nil {
		log.Debugf("Deploy stroageclass %s", name)
		yaml, err := StorageClassYaml(id, name)
		if err != nil {
			return err
		}
		if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
			return err
		}
	}
	return nil
}

func Stop(cli client.Client, id, name string) error {
	log.Debugf("Undeploy stroageclass %s", name)
	yaml, err := StorageClassYaml(id, name)
	if err != nil {
		return err
	}
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil {
		return err
	}

	log.Debugf("Undeploy fscsi")
	yaml, err = fscsiYaml()
	if err != nil {
		return err
	}
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	log.Debugf("Undeploy csi-cfg")
	yaml, err = CSICfgYaml("", "")
	if err != nil {
		return err
	}
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	return nil
}
