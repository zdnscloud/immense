package osd

import (
	"fmt"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	cephclient "github.com/zdnscloud/immense/pkg/ceph/client"
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/ceph/util"
	"github.com/zdnscloud/immense/pkg/ceph/zap"
	"github.com/zdnscloud/immense/pkg/common"
	"strings"
	"time"
)

func Start(cli client.Client, fsid, host, dev string, monsvc map[string]string) error {
	var monEndpoints []string
	for _, ip := range monsvc {
		ep := "v1:" + ip + ":" + global.MonPortV1
		monEndpoints = append(monEndpoints, ep)
	}
	eps := strings.Replace(strings.Trim(fmt.Sprint(monEndpoints), "[]"), " ", ",", -1)
	members := strings.Replace(strings.Trim(fmt.Sprint(global.MonMembers), "[]"), " ", ",", -1)
	log.Debugf("Deploy osd %s:%s", host, dev)
	yaml, err := osdYaml(fsid, host, dev, members, eps)
	if err != nil {
		return err
	}
	if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	waitDone(cli, host, dev)
	return nil
}

func Stop(cli client.Client, host, dev string) error {
	log.Debugf("Undeploy osd %s:%s", host, dev)
	var fsid, members, eps string
	yaml, err := osdYaml(fsid, host, dev, members, eps)
	if err != nil {
		return err
	}
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	checkDone(cli, host, dev)
	if err := zap.Do(cli, host, dev); err != nil {
		return err
	}
	return remove(host)
}

func waitDone(cli client.Client, host, dev string) {
	log.Debugf("Wait osd running %s:%s, this will take some time", host, dev)
	name := "ceph-osd-" + host + "-" + dev
	var done bool
	for !done {
		time.Sleep(10 * time.Second)
		if !common.IsDsReady(cli, common.StorageNamespace, name) {
			continue
		}
		done = true
	}
}

func checkDone(cli client.Client, host, dev string) {
	log.Debugf("Wait osd end %s:%s", host, dev)
	name := "ceph-osd-" + host + "-" + dev
	var ready bool
	for !ready {
		time.Sleep(10 * time.Second)
		del, err := util.CheckPodDel(cli, name)
		if err != nil || !del {
			continue
		}
		ready = true
	}
}

func remove(host string) error {
	name := "ceph-osd-" + host + "-"
	ids, err := cephclient.GetHostToIDs(name)
	if err != nil {
		return err
	}
	for _, id := range ids {
		log.Debugf("[ceph] Remove osd %s from ceph cluster", id)
		osdName := "osd." + id
		if err := cephclient.OutOsd(id); err != nil {
			log.Warnf("[ceph] Ceph out osd %s failed , err:%s", id, err.Error())
		}
		if err := cephclient.RemoveCrush(osdName); err != nil {
			log.Warnf("[ceph] Ceph osd remove crush %s failed , err:%s", osdName, err.Error())
		}
		if err := cephclient.RmOsd(id); err != nil {
			log.Warnf("[ceph] Ceph osd rm %s failed , err:%s", id, err.Error())
		}
		if err := cephclient.RmOsdAuth(osdName); err != nil {
			log.Warnf("[ceph] Ceph auth del osd.%s failed , err:%s", id, err.Error())
		}
	}
	return nil
}
