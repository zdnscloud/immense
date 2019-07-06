package handle

import (
	"context"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	cephclient "github.com/zdnscloud/immense/pkg/ceph/client"
	"github.com/zdnscloud/immense/pkg/common"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"strconv"
	"strings"
	"time"
)

func StatusControl(cli client.Client, name string) {
	log.Debugf("[ceph-status-controller] Start")
	for {
		time.Sleep(60 * time.Second)
		storagecluster := storagev1.Cluster{}
		err := cli.Get(context.TODO(), k8stypes.NamespacedName{common.StorageNamespace, name}, &storagecluster)
		if err != nil {
			log.Warnf("[ceph-status-controller] Get storage cluster %s failed, err:%s", name, err.Error())
			log.Debugf("[ceph-status-controller] Stop")
			return
		}
		state, message, capacity, err := getInfo(storagecluster)
		if err != nil {
			log.Warnf("[ceph-status-controller] Get ceph status failed, err:%s", err.Error())
			continue
		}
		storagecluster.Status.State = state
		storagecluster.Status.Message = message
		storagecluster.Status.Capacity = capacity
		log.Debugf("[ceph-status-controller] Update storage cluster %s", name)
		err = cli.Update(context.TODO(), &storagecluster)
		if err != nil {
			log.Warnf("[ceph-status-controller] Update storage cluster %s failed, err:%s", name, err.Error())
			continue
		}
	}
}

func getInfo(storagecluster storagev1.Cluster) (string, string, storagev1.Capacity, error) {
	var state, message string
	var capacity storagev1.Capacity
	out, err := cephclient.CheckHealth()
	if err != nil {
		return state, message, capacity, err
	}
	if strings.Contains(out, "HEALTH_OK") {
		state = "HEALTH_OK"
	} else if strings.Contains(out, "HEALTH_WARN") {
		state = "HEALTH_WARN"
	} else {
		state = "HEALTH_ERR"
	}
	message = out
	infos, err := cephclient.GetDF()
	if err != nil {
		return state, message, capacity, err
	}
	summary := infos.Summary
	unit := int64(1024)
	capacity.Total = storagev1.Size{
		Total: strconv.FormatInt(summary.Total*unit, 10),
		Used:  strconv.FormatInt(summary.Used*unit, 10),
		Free:  strconv.FormatInt(summary.Avail*unit, 10),
	}
	nodes := infos.Nodes
	instances := make([]storagev1.Instance, 0)
	for _, n := range nodes {
		name, err := cephclient.GetIDToHost(strconv.FormatInt(n.ID, 10))
		if err != nil {
			continue
		}
		host, dev := osdSplit(storagecluster, name)
		info := storagev1.Instance{
			Host: host,
			Dev:  dev,
			Stat: true,
			Info: storagev1.Size{
				Total: strconv.FormatInt(n.Total*unit, 10),
				Used:  strconv.FormatInt(n.Used*unit, 10),
				Free:  strconv.FormatInt(n.Avail*unit, 10),
			},
		}
		instances = append(instances, info)
	}
	capacity.Instances = instances
	return state, message, capacity, nil
}

func osdSplit(storagecluster storagev1.Cluster, podname string) (string, string) {
	for _, h := range storagecluster.Spec.Hosts {
		for _, d := range h.BlockDevices {
			if strings.Contains(podname, h.NodeName) && strings.Contains(podname, d[5:]) {
				return h.NodeName, d
			}
		}
	}
	return "", ""
}
