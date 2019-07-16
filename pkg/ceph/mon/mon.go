package mon

import (
	"context"
	"errors"
	"fmt"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	cephclient "github.com/zdnscloud/immense/pkg/ceph/client"
	"github.com/zdnscloud/immense/pkg/ceph/fscsi"
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/ceph/util"
	"github.com/zdnscloud/immense/pkg/common"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
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
	log.Debugf("Wait all mon running")
	for i := 0; i < 60; i++ {
		mons, err := util.GetMonIPs(cli)
		if err != nil {
			return false, err
		}
		if len(mons) == global.MonNum {
			return true, nil
		}
		time.Sleep(5 * time.Second)
	}
	return false, errors.New("Timeout. Mon has not ready")
}

func Watch(cli client.Client) {
	log.Debugf("[ceph-mon-watcher] Start")
	for {
		time.Sleep(60 * time.Second)
		if !cephclient.CheckConf() {
			log.Debugf("[ceph-mon-watcher] Stop")
			return
		}
		ep := corev1.Endpoints{}
		err := cli.Get(context.TODO(), k8stypes.NamespacedName{common.StorageNamespace, global.MonSvc}, &ep)
		if err != nil {
			log.Warnf("[ceph-mon-watcher] Get endpoints %s failed. Err:%s", global.MonSvc, err.Error())
			continue
		}
		svc := make(map[string]string)
		for _, sub := range ep.Subsets {
			for _, ads := range sub.Addresses {
				svc[ads.Hostname] = ads.IP
			}
		}
		moninfo, err := cephclient.GetMon()
		if err != nil {
			log.Warnf("[ceph-mon-watcher] Get mon dump failed. Err:%s", err.Error())
			continue
		}
		unnormal := make([]string, 0)
		for _, mon := range moninfo.Mons {
			v, ok := svc[mon.Name]
			if !ok || v != strings.Split(mon.Addr, ":")[0] {
				unnormal = append(unnormal, mon.Name)
			}
		}
		if len(unnormal) == 0 {
			continue
		}
		log.Debugf("[ceph-mon-watcher] Watch mon %s has unnormal, remove it now", unnormal)
		for _, name := range unnormal {
			err := cephclient.Rmmon(name)
			if err != nil {
				log.Warnf("[ceph-mon-watcher] Remove mon %s failed. Err:%s", name, err.Error())
				continue
			}
		}
		var mons []string
		for _, v := range svc {
			mons = append(mons, v+":6789")
		}
		log.Debugf("[ceph-mon-watcher] Watch mons has changed, redeploy storageclass now")
		monitors := strings.Replace(strings.Trim(fmt.Sprint(mons), "[]"), " ", ",", -1)
		if err := redeployStorageclass(cli, monitors); err != nil {
			log.Warnf("[ceph-mon-watcher] Redeploy storageclass %s failed. Err:%s", global.StorageClassName, err.Error())
			continue
		}
	}
}

func redeployStorageclass(cli client.Client, monitors string) error {
	sc := storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{Name: global.StorageClassName, Namespace: ""},
	}
	if err := cli.Delete(context.TODO(), &sc); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	yaml, err := fscsi.StorageClassYaml(monitors)
	if err != nil {
		return err
	}
	if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
		return err
	}
	return nil
}
