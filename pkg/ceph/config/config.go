package config

import (
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/ceph/util"
	"github.com/zdnscloud/immense/pkg/common"
)

func Start(cli client.Client, uuid, networks, adminkey, monkey string, number int) error {
	log.Debugf("Deploy service %s", global.MonSvc)
	yaml, err := svcYaml()
	if err != nil {
		return err
	}
	if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
		return err
	}

	exist, err := util.CheckConfigMap(cli, common.StorageNamespace, global.ConfigMapName)
	if !exist || err != nil {
		log.Debugf("Deploy configmap %s", global.ConfigMapName)
		yaml, err = confYaml(uuid, networks, adminkey, monkey, number)
		if err != nil {
			return err
		}
		if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
			return err
		}
	}

	exist, err = util.CheckSecret(cli, common.StorageNamespace, global.SecretName)
	if !exist || err != nil {
		log.Debugf("Deploy secret %s", global.SecretName)
		yaml, err = secretYaml("admin", adminkey)
		if err != nil {
			return err
		}
		if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
			return err
		}
	}
	log.Debugf("Deploy serviceaccount %s", global.ServiceAccountName)
	yaml, err = saYaml()
	if err != nil {
		return err
	}
	if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	return nil
}

func Stop(cli client.Client, uuid, networks, adminkey, monkey string, number int) error {
	log.Debugf("Undeploy service %s", global.MonSvc)
	yaml, err := svcYaml()
	if err != nil {
		return err
	}
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil {
		return err
	}

	log.Debugf("Undeploy configmap %s", global.ConfigMapName)
	yaml, err = confYaml(uuid, networks, adminkey, monkey, number)
	if err != nil {
		return err
	}
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil {
		return err
	}

	log.Debugf("Undeploy secret %s", global.ConfigMapName)
	yaml, err = secretYaml("", "")
	if err != nil {
		return err
	}
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	log.Debugf("Undeploy serviceaccount %s", global.ServiceAccountName)
	yaml, err = saYaml()
	if err != nil {
		return err
	}
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	return nil
}
