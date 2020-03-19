package prepare

import (
	"fmt"
	"strings"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	"github.com/zdnscloud/immense/pkg/common"
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
	name := "ceph-job-prepare-" + host
	return common.WaitPodSucceeded(cli, common.StorageNamespace, name)
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
