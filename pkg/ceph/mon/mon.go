package mon

import (
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	cephclient "github.com/zdnscloud/immense/pkg/ceph/client"
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/ceph/util"
	"github.com/zdnscloud/immense/pkg/common"
	"strings"
	"time"
)

func Start(cli client.Client, networks string) error {
	log.Debugf("Deploy mon")
	yaml, err := monYaml(networks)
	if err != nil {
		return err
	}
	if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	check(cli)
	return nil
}

func Stop(cli client.Client, networks string) error {
	log.Debugf("Undeploy mon")
	yaml, err := monYaml(networks)
	if err != nil {
		return err
	}
	return helper.DeleteResourceFromYaml(cli, yaml)
}

func check(cli client.Client) {
	log.Debugf("Wait all mon running, this will take some time")
	var ready bool
	for !ready {
		time.Sleep(10 * time.Second)
		if !isReady(cli) {
			continue
		}
		if !isHealth() {
			continue
		}
		ready = true
	}
}

func isHealth() bool {
	message, err := cephclient.CheckHealth()
	if err != nil || !strings.Contains(message, "HEALTH_OK") {
		log.Warnf("Ceph cluster is not yet healthy, try again later")
		return false
	}
	return true
}

func isReady(cli client.Client) bool {
	num, err := util.GetMonDpReadyNum(cli, common.StorageNamespace, "ceph-mon")
	if err != nil || num != global.MonNum {
		log.Warnf("Mons has not all running, try again later")
		return false
	}
	return true
}
