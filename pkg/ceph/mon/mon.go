package mon

import (
	"context"
	"errors"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	cephclient "github.com/zdnscloud/immense/pkg/ceph/client"
	"github.com/zdnscloud/immense/pkg/ceph/fscsi"
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/ceph/util"
	"github.com/zdnscloud/immense/pkg/common"
	corev1 "k8s.io/api/core/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"net"
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
	ready, err := check(cli)
	if err != nil {
		return err
	}
	if !ready {
		return errors.New("Timeout. Ceph cluster has not ready")
	}
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

func check(cli client.Client) (bool, error) {
	log.Debugf("Wait all mon running, this will take some time")
	var mons []string
	for i := 0; i < 60; i++ {
		time.Sleep(5 * time.Second)
		if len(mons) != global.MonNum {
			svc, err := util.GetMonSvc(cli)
			if err != nil || len(svc) != global.MonNum {
				continue
			}
			mons = append(mons, svc...)
		}
		if checkMonConnect(mons) {
			return true, nil
		}
		if isHealth() {
			return true, nil
		}
	}
	return false, errors.New("Timeout. Mons has not ready")
}

func isHealth() bool {
	message, err := cephclient.CheckHealth()
	if err != nil || !strings.Contains(message, "HEALTH_OK") {
		log.Warnf("Ceph cluster is not yet healthy, try again later")
		return false
	}
	return true
}

func checkMonConnect(ips []string) bool {
	connTimeout := 5 * time.Second
	for _, ip := range ips {
		addr := ip + ":" + global.MonPort
		_, err := net.DialTimeout("tcp", addr, connTimeout)
		if err != nil {
			log.Warnf("Mon %s can not connection, try again later", addr)
			return false
		}
	}
	return true
}

func Watch(cli client.Client, uuid string) {
	log.Debugf("[ceph-mon-watcher] Start")
	for {
		time.Sleep(60 * time.Second)
		if !cephclient.CheckConf() {
			log.Debugf("[ceph-mon-watcher] Stop")
			return
		}
		svc, err := getMonEp(cli)
		if err != nil {
			log.Warnf("[ceph-mon-watcher] Get endpoints %s failed. Err: %s. skip", global.MonSvc, err.Error())
			continue
		}
		unnormal, err := getUnnormalMon(svc)
		if err != nil {
			log.Warnf("[ceph-mon-watcher] Get mon dump failed. Err: %s. skip", err.Error())
			continue
		}
		if len(unnormal) == 0 {
			continue
		}
		log.Debugf("[ceph-mon-watcher] Watch mon %s has unnormal, remove it now", unnormal)
		for _, name := range unnormal {
			if err := cephclient.Rmmon(name); err != nil {
				log.Warnf("[ceph-mon-watcher] Remove mon %s failed. Err:%s. skip", name, err.Error())
				continue
			}
		}
		log.Debugf("[ceph-mon-watcher] Watch mons has changed, update csi configmap %s now", global.CSIConfigmapName)
		if err := redeployCSICfg(cli, svc, uuid); err != nil {
			log.Warnf("[ceph-mon-watcher] Redeploy csi configmap %s failed. Err:%s. skip", global.CSIConfigmapName, err.Error())
			continue
		}
	}
}

func redeployCSICfg(cli client.Client, svc map[string]string, uuid string) error {
	var mons string
	for _, ip := range svc {
		mon := "\"" + ip + ":" + global.MonPort + "\","
		mons += mon
	}
	ms := strings.TrimRight(mons, ",")
	yaml, err := fscsi.CSICfgYaml(uuid, ms)
	if err != nil {
		return err
	}
	if err := helper.UpdateResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	return nil
}

func getMonEp(cli client.Client) (map[string]string, error) {
	svc := make(map[string]string)
	ep := corev1.Endpoints{}
	if err := cli.Get(context.TODO(), k8stypes.NamespacedName{common.StorageNamespace, global.MonSvc}, &ep); err != nil {
		return svc, err
	}
	for _, sub := range ep.Subsets {
		for _, ads := range sub.Addresses {
			svc[ads.TargetRef.Name] = ads.IP
		}
	}
	return svc, nil
}

func getUnnormalMon(svc map[string]string) ([]string, error) {
	unnormal := make([]string, 0)
	moninfo, err := cephclient.GetMon()
	if err != nil {
		return unnormal, err
	}
	for _, mon := range moninfo.Mons {
		v, ok := svc[mon.Name]
		if !ok || v != strings.Split(mon.Addr, ":")[0] {
			unnormal = append(unnormal, mon.Name)
		}
	}
	return unnormal, nil
}
