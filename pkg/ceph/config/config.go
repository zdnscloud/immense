package config

import (
	"fmt"
	"strings"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/ceph/util"
	"github.com/zdnscloud/immense/pkg/common"
)

func Start(cli client.Client, uuid, adminkey, monkey string, number int) error {
	log.Debugf("Deploy service %s", global.MonSvc)
	for _, id := range global.MonMembers {
		yaml, err := svcYaml(id)
		if err != nil {
			return err
		}
		if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
			return err
		}
	}

	monsvc, err := util.GetMonSvcMap(cli)
	if err != nil {
		return err
	}
	var monEndpoints []string
	for _, ip := range monsvc {
		ep := "v1:" + ip + ":" + global.MonPortV1
		monEndpoints = append(monEndpoints, ep)
	}
	eps := strings.Replace(strings.Trim(fmt.Sprint(monEndpoints), "[]"), " ", ",", -1)
	members := strings.Replace(strings.Trim(fmt.Sprint(global.MonMembers), "[]"), " ", ",", -1)

	exist, err := util.CheckConfigMap(cli, common.StorageNamespace, global.ConfigMapName)
	if !exist || err != nil {
		log.Debugf("Deploy configmap %s", global.ConfigMapName)
		yaml, err := confYaml(uuid, adminkey, monkey, eps, members, number)
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
		yaml, err := secretYaml("admin", adminkey)
		if err != nil {
			return err
		}
		if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
			return err
		}
	}
	log.Debugf("Deploy serviceaccount %s", global.ServiceAccountName)
	yaml, err := saYaml()
	if err != nil {
		return err
	}
	if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	return nil
}

func Stop(cli client.Client, uuid, adminkey, monkey string, number int) error {
	log.Debugf(fmt.Sprintf("Undeploy service %s", global.MonSvc))
	for _, id := range global.MonMembers {
		yaml, err := svcYaml(id)
		if err != nil {
			return err
		}
		if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil {
			return err
		}
	}
	var eps, members string
	log.Debugf("Undeploy configmap %s", global.ConfigMapName)
	yaml, err := confYaml(uuid, adminkey, monkey, eps, members, number)
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
