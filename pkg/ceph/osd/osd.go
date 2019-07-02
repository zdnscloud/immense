package osd

import (
	"errors"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	cephclient "github.com/zdnscloud/immense/pkg/ceph/client"
	"github.com/zdnscloud/immense/pkg/ceph/util"
	"github.com/zdnscloud/immense/pkg/ceph/zap"
	"strings"
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

func Remove(cli client.Client, host, dev string) error {
	log.Debugf("Ceph del host:%s dev:%s", host, dev)
	name := "ceph-osd-" + host + "-" + dev
	ip, err := util.GetPodIp(cli, name)
	if err != nil {
		return err
	}
	id, err := cephclient.GetOsdID(ip)
	if err != nil {
		return err
	}
	log.Debugf("Ceph get osd, id is:%s", id)
	osdName := "osd." + id
	log.Debugf("Ceph osd crush reweight %s", osdName)
	if err := cephclient.ReweigtOsd(osdName); err != nil {
		return err
	}
	time.Sleep(60 * time.Second)
	if err := checkHealth(); err != nil {
		return err
	}
	log.Debugf("Ceph osd out id %s", id)
	if err := cephclient.OutOsd(id); err != nil {
		return err
	}
	if err := checkHealth(); err != nil {
		return err
	}
	log.Debugf("Ceph osd crush remove %s", osdName)
	if err := cephclient.RemoveCrush(osdName); err != nil {
		return err
	}
	if err := Stop(cli, host, dev); err != nil {
		return err
	}
	if err := checkHealth(); err != nil {
		return err
	}
	log.Debugf("Ceph osd rm %s", id)
	if err := cephclient.RmOsd(id); err != nil {
		return err
	}
	if err := checkHealth(); err != nil {
		return err
	}
	log.Debugf("Ceph auth del %s", osdName)
	if err := cephclient.RmOsdAuth(osdName); err != nil {
		return err
	}
	return nil
}

func checkHealth() error {
	log.Debugf("Wait Ceph health ok, it will check 10 times")
	//time.Sleep(60 * time.Second)
	var out string
	for i := 0; i < 10; i++ {
		out, err := cephclient.CheckHealth()
		if err != nil {
			return err
		}
		log.Debugf("[%d], %s", i+1, out)
		if strings.Contains(out, "HEALTH_OK") {
			return nil
		}
		if strings.Contains(out, "HEALTH_ERR") {
			return errors.New("Ceph cluster is ERR")
		}
		time.Sleep(60 * time.Second)
	}
	log.Warnf("Wait Timeout. Ceph status %s", out)
	return errors.New("Ceph cluster is not normal")
}
