package osd

import (
	"errors"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	cephclient "github.com/zdnscloud/immense/pkg/ceph/client"
	"github.com/zdnscloud/immense/pkg/ceph/util"
	"github.com/zdnscloud/immense/pkg/ceph/zap"
	"time"
)

func Start(cli client.Client, host, dev string) error {
	name := "ceph-osd-" + host + "-" + dev
	ip, err := util.GetPodIp(cli, name)
	if err != nil {
		return err
	}
	if ip == "" {
		if err := zap.Do(cli, host, dev); err != nil {
			return err
		}
	}

	log.Debugf("Deploy osd %s:%s", host, dev)
	yaml, err := osdYaml(host, dev)
	if err != nil {
		return err
	}
	return helper.CreateResourceFromYaml(cli, yaml)
}

func Stop(cli client.Client, host, dev string) error {
	log.Debugf("Undeploy osd %s:%s", host, dev)
	yaml, err := osdYaml(host, dev)
	if err != nil {
		return err
	}
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	del, err := check(cli, host, dev)
	if err != nil {
		return err
	}
	if !del {
		return errors.New("Osd process is not clean up!")
	}
	return zap.Do(cli, host, dev)
}

func check(cli client.Client, host, dev string) (bool, error) {
	log.Debugf("Wait osd end %s:%s", host, dev)
	name := "ceph-osd-" + host + "-" + dev
	for i := 0; i < 60; i++ {
		del, err := util.IsPodDel(cli, name)
		if err != nil {
			return false, err
		}
		if del {
			return true, nil
		}
		time.Sleep(5 * time.Second)
	}
	return false, nil
}

func Watch() {
	log.Debugf("[ceph-osd-watcher] Start")
	for {
		if !cephclient.CheckConf() {
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
			log.Warnf("[ceph-osd-watcher] Get osd in and out failed , err:%s. skip", err.Error())
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
