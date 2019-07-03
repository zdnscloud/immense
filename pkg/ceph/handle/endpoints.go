package handle

import (
	"context"
	"fmt"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	"github.com/zdnscloud/immense/pkg/ceph/fscsi"
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/ceph/mon"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"reflect"
	"strings"
)

func DoEndpointsUpdate(cli client.Client, oldendpoint, newendpoint *corev1.Endpoints) error {
	if newendpoint.Name != global.MonSvc {
		return nil
	}

	sc := storagev1.StorageClass{}
	err := cli.Get(context.TODO(), k8stypes.NamespacedName{"", global.StorageClassName}, &sc)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	nodename, monitors, change := getNodeFromEndpoint(oldendpoint, newendpoint)
	if !change {
		return nil
	}
	if monitors != "" {
		log.Debugf("Watch k8s endpoint %s has changed, redeploy storageclass %s now", global.MonSvc, global.StorageClassName)
		if err = cli.Delete(context.TODO(), &sc); err != nil {
			return err
		}
		yaml, err := fscsi.StorageClassYaml(monitors)
		if err != nil {
			return err
		}
		if err := helper.CreateResourceFromYaml(cli, yaml); err != nil {
			return err
		}
	}
	if nodename != "" {
		log.Debugf("Mon on %s has delete, remove it from ceph cluster now", nodename)
		if err := mon.Remove(nodename); err != nil {
			return err
		}
	}
	return nil
}

func getNodeFromEndpoint(oldendpoint, newendpoint *corev1.Endpoints) (string, string, bool) {
	oldips := make(map[string]string)
	oldslice := make([]string, 0)
	for _, sub := range oldendpoint.Subsets {
		for _, ads := range sub.Addresses {
			oldips[ads.IP] = *ads.NodeName
			oldslice = append(oldslice, oldips[ads.IP])
		}
	}
	newips := make(map[string]string)
	newslice := make([]string, 0)
	for _, sub := range newendpoint.Subsets {
		for _, ads := range sub.Addresses {
			newips[ads.IP] = *ads.NodeName
			newslice = append(newslice, newips[ads.IP])
		}
	}
	if reflect.DeepEqual(oldslice, newslice) {
		return "", "", false
	}
	var nodename string
	for k, v := range oldips {
		_, ok := newips[k]
		if !ok {
			nodename = v
			break
		}
	}
	var mons []string
	for k := range newips {
		mons = append(mons, k+":6789")
	}
	monitors := strings.Replace(strings.Trim(fmt.Sprint(mons), "[]"), " ", ",", -1)
	return nodename, monitors, true
}
