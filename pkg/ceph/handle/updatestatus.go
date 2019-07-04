package handle

import (
	"context"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	cephclient "github.com/zdnscloud/immense/pkg/ceph/client"
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/common"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"math"
	"strconv"
	"strings"
	"time"
)

func StatusControl(cli client.Client, name string) {
	log.Debugf("[status-controller] Start")
	for {
		if !cephclient.CheckConf() {
			log.Debugf("[status-controller] Stop")
			return
		}
		time.Sleep(60 * time.Second)
		storagecluster := storagev1.Cluster{}
		err := cli.Get(context.TODO(), k8stypes.NamespacedName{common.StorageNamespace, name}, &storagecluster)
		if err != nil {
			log.Warnf("[status-controller] Get storage cluster %s failed, err:%s", name, err.Error())
			continue
		}
		state, message, capacity, err := getInfo()
		if err != nil {
			log.Warnf("[status-controller] Get ceph status failed, err:%s", err.Error())
			continue
		}
		storagecluster.Status.State = state
		storagecluster.Status.Message = message
		storagecluster.Status.Capacity = capacity
		log.Debugf("[status-controller] Update storage cluster %s", name)
		err = cli.Update(context.TODO(), &storagecluster)
		if err != nil {
			log.Warnf("[status-controller] Update storage cluster %s failed, err:%s", name, err.Error())
			continue
		}
	}
}

func getInfo() (string, string, map[string]string, error) {
	var state, message string
	capacity := make(map[string]string)
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
	size, err := cephclient.GetFSDF(global.CephFsDate)
	if err != nil {
		return state, message, capacity, err
	}
	tmp := strings.Fields(size)
	used := sizeswitch(tmp[2], tmp[3])
	free := sizeswitch(tmp[5], tmp[6])
	total := used + free
	capacity["total"] = strconv.FormatFloat(total, 'f', -1, 64)
	capacity["used"] = strconv.FormatFloat(used, 'f', -1, 64)
	capacity["free"] = strconv.FormatFloat(free, 'f', -1, 64)
	return state, message, capacity, nil
}

func sizeswitch(size, unit string) float64 {
	var num int64
	switch unit {
	case "GiB":
		num = 1
	case "MiB":
		num = 1024
	case "KiB":
		num = 1024 * 1024
	case "B":
		num = 1024 * 1024 * 1024
	}
	f, _ := strconv.ParseFloat(size, 64)
	return math.Floor(f/float64(num) + 0/5)
}
