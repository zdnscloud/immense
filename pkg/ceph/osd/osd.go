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

	name := "ceph-osd-" + host + "-" + dev
	ok, err := util.CheckPodPhase(cli, name, "Running")
	if err != nil {
		return err
	}
	if !ok {
		if err := zap.Do(cli, host, dev); err != nil {
			return err
		}
	}
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
	return zap.Do(cli, host, dev)
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

func Watch() {
	log.Debugf("[ceph-osd-watcher] Start")
	for {
		if !util.CheckConf() {
			log.Debugf("[ceph-osd-watcher] Stop")
			return
		}
		time.Sleep(40 * time.Second)
		downandin, err := cephclient.GetDownOsdIDs("in")
		if err != nil {
			log.Warnf("[ceph-osd-watcher] Get osd down and in failed , err:%s", err.Error())
			continue
		}
		for _, id := range downandin {
			log.Debugf("[ceph-osd-watcher] Ceph osd out id %s", id)
			if err := cephclient.OutOsd(id); err != nil {
				log.Warnf("[ceph-osd-watcher] Ceph out osd %s failed , err:%s", id, err.Error())
				break
			}
		}
		time.Sleep(20 * time.Second)
		downandout, err := cephclient.GetDownOsdIDs("out")
		if err != nil {
			log.Warnf("[ceph-osd-watcher] Get osd down and out failed , err:%s. skip", err.Error())
			continue
		}
		for _, id := range downandout {
			osdName := "osd." + id
			log.Debugf("[ceph-osd-watcher] Ceph osd crush remove %s", osdName)
			if err := cephclient.RemoveCrush(osdName); err != nil {
				log.Warnf("[ceph-osd-watcher] Ceph osd remove crush %s failed , err:%s", osdName, err.Error())
				break
			}
			log.Debugf("[ceph-osd-watcher] Ceph osd rm %s", id)
			if err := cephclient.RmOsd(id); err != nil {
				log.Warnf("[ceph-osd-watcher] Ceph osd rm %s failed , err:%s", id, err.Error())
				break
			}
			log.Debugf("[ceph-osd-watcher] Ceph auth del %s", osdName)
			if err := cephclient.RmOsdAuth(osdName); err != nil {
				log.Warnf("[ceph-osd-watcher] Ceph auth del osd.%s failed , err:%s", id, err.Error())
				break
			}
		}
		upandout, err := cephclient.GetUpAndOutOsdIDs()
		if err != nil {
			log.Warnf("[ceph-osd-watcher] Get osd up and out failed , err:%s. skip", err.Error())
			continue
		}
		for _, id := range upandout {
			log.Debugf("[ceph-osd-watcher] Ceph osd in id %s", id)
			if err := cephclient.InOsd(id); err != nil {
				log.Warnf("[ceph-osd-watcher] Ceph in osd %s failed , err:%s", id, err.Error())
				break
			}
		}
	}
}
