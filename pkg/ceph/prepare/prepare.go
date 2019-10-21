package prepare

import (
	"fmt"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	"github.com/zdnscloud/immense/pkg/ceph/util"
	"strings"
	"time"
)

func Do(cli client.Client, host string, devs []string) error {
	devinfo := strings.Trim(fmt.Sprint(devs), "[]")
	log.Debugf("Prepare host %s device %s", host, devinfo)
	yaml, err := prepareYaml(host, devinfo)
	if err != nil {
		return err
	}
	if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	check(cli, host)
	return nil
}

func check(cli client.Client, host string) {
	log.Debugf("Wait prepare done %s", host)
	name := "ceph-job-prepare-" + host
	var ready bool
	for !ready {
		time.Sleep(10 * time.Second)
		ok, err := util.CheckPodPhase(cli, name, "Succeeded")
		if err != nil || !ok {
			continue
		}
		ready = true
	}
}

func Delete(cli client.Client, host string) error {
	log.Debugf("Undeploy Prepare host %s", host)
	var devinfo string
	yaml, err := prepareYaml(host, devinfo)
	if err != nil {
		return err
	}
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	return nil
}
