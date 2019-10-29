package mon

import (
	"fmt"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/ceph/util"
	"strings"
)

func Start(cli client.Client, fsid string, monsvc map[string]string) error {
	monHosts := util.GetMonHosts(monsvc)
	monMembers := strings.Replace(strings.Trim(fmt.Sprint(global.MonMembers), "[]"), " ", ",", -1)
	for _, id := range global.MonMembers {
		log.Debugf(fmt.Sprintf("Deploy mon %s", id))
		addr := monsvc[id]
		yaml, err := monYaml(fsid, id, addr, monHosts, monMembers)
		if err != nil {
			return err
		}
		if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
			return err
		}
		name := global.MonDpName + "-" + id
		util.WaitDpReady(cli, name)
	}
	return nil
}

func Stop(cli client.Client) error {
	mons := []string{"a", "b", "c"}
	for _, id := range mons {
		log.Debugf(fmt.Sprintf("Undeploy mon %s", id))
		var fsid, addr, monHosts, monMembers string
		yaml, err := monYaml(fsid, id, addr, monHosts, monMembers)
		if err != nil {
			return err
		}
		if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil {
			return err
		}
	}
	return nil
}
