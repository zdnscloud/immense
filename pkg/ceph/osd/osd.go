package osd

import (
	"fmt"
	"strings"
	"time"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	cephclient "github.com/zdnscloud/immense/pkg/ceph/client"
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/ceph/util"
	"github.com/zdnscloud/immense/pkg/ceph/zap"
	"github.com/zdnscloud/immense/pkg/common"
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
	waitOsdRunning(cli, host, dev)
	return nil
}

func Remove(cli client.Client, host, dev string) error {
	log.Debugf("Undeploy osd %s:%s", host, dev)
	var fsid, members, eps string
	yaml, err := osdYaml(fsid, host, dev, members, eps)
	if err != nil {
		return err
	}
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	waitOsdDelete(cli, host, dev)
	if err := zap.Do(cli, host, dev); err != nil {
		return err
	}
	return nil
}

func Stop(cli client.Client, hostInfos []storagev1.HostInfo) error {
	var allIds []string
	for _, hostInfo := range hostInfos {
		name := "ceph-osd-" + hostInfo.NodeName + "-"
		ids, err := cephclient.GetHostToIDs(name)
		if err != nil {
			return err
		}
		allIds = append(allIds, ids...)
	}
	if err := reweight(allIds); err != nil {
		return err
	}
	waitRebalance()
	if err := out(allIds); err != nil {
		return err
	}
	for _, hostInfo := range hostInfos {
		for _, dev := range hostInfo.BlockDevices {
			if err := Remove(cli, hostInfo.NodeName, dev[5:]); err != nil {
				return err
			}
		}
	}
	if err := remove(allIds); err != nil {
		return err
	}
	return nil
}

func waitOsdRunning(cli client.Client, host, dev string) {
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

func waitOsdDelete(cli client.Client, host, dev string) {
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

func waitRebalance() {
	log.Debugf("Wait ceph pgs rebalance to finish, this will take some time")
	var finish bool
	for !finish {
		time.Sleep(60 * time.Second)
		message, err := cephclient.CheckHealth()
		if err != nil || !strings.Contains(message, "HEALTH_OK") {
			continue
		}
		finish = true
	}
}

func reweight(ids []string) error {
	for _, id := range ids {
		log.Debugf("[ceph] Ceph reweight osd %s to 0 from ceph cluster", id)
		osdName := "osd." + id
		if err := cephclient.ReweigtOsd(osdName); err != nil {
			return err
			log.Warnf("[ceph] Ceph reweight osd %s failed , err:%s", id, err.Error())
		}
	}
	return nil
}

func out(ids []string) error {
	for _, id := range ids {
		log.Debugf("[ceph] Ceph out osd %s from ceph cluster", id)
		if err := cephclient.OutOsd(id); err != nil {
			log.Warnf("[ceph] Ceph out osd %s failed , err:%s", id, err.Error())
		}
	}
	return nil
}

func remove(ids []string) error {
	for _, id := range ids {
		log.Debugf("[ceph] Remove osd %s from ceph cluster", id)
		osdName := "osd." + id
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
