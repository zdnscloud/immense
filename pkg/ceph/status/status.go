package status

import (
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	cephclient "github.com/zdnscloud/immense/pkg/ceph/client"
	"github.com/zdnscloud/immense/pkg/common"
	"strconv"
	"strings"
	"time"
)

func Watch(cli client.Client, name string) {
	log.Debugf("[ceph-status-controller] Start")
	for {
		time.Sleep(60 * time.Second)
		storagecluster, err := common.GetStorage(cli, name)
		if err != nil {
			log.Warnf("[ceph-status-controller] Get storage %s config with blocks failed. Err: %s", name, err.Error())
			log.Debugf("[ceph-status-controller] Stop")
			return
		}
		if storagecluster.Status == "Updating" || storagecluster.Status == "Creating" {
			continue
		}
		status, err := getStatus(storagecluster)
		if err != nil {
			log.Warnf("[ceph-status-controller] Get ceph status failed. Err: %s", err.Error())
			continue
		}
		if err := common.UpdateStatus(cli, name, status); err != nil {
			log.Warnf("[ceph-status-controller] Update storage cluster %s failed. Err: %s", name, err.Error())
		}
	}
}

func getStatus(storagecluster common.Storage) (storagev1.ClusterStatus, error) {
	var status storagev1.ClusterStatus
	phase, message, err := getPhaseAndMsg()
	if err != nil {
		return status, err
	}
	capacity, err := getCapacity(storagecluster)
	if err != nil {
		return status, err
	}
	status.Phase = phase
	status.Message = message
	status.Capacity = capacity
	return status, nil
}

func getPhaseAndMsg() (string, string, error) {
	var phase, message string
	message, err := cephclient.CheckHealth()
	if err != nil {
		return phase, message, err
	}
	if strings.Contains(message, "HEALTH_OK") {
		phase = "Running"
	} else if strings.Contains(message, "HEALTH_WARN") {
		phase = "Warnning"
	} else {
		phase = "Error"
	}
	return phase, message, nil
}

var unit = int64(1024)

func getCapacity(storagecluster common.Storage) (storagev1.Capacity, error) {
	var capacity storagev1.Capacity
	infos, err := cephclient.GetDF()
	if err != nil {
		return capacity, err
	}
	capacity.Total = storagev1.Size{
		Total: strconv.FormatInt(infos.Summary.Total*unit, 10),
		Used:  strconv.FormatInt(infos.Summary.Used*unit, 10),
		Free:  strconv.FormatInt(infos.Summary.Avail*unit, 10),
	}
	online := getOnlineInstances(storagecluster, infos.Nodes)
	offline := getOfflineInstances(storagecluster, online)
	capacity.Instances = append(online, offline...)
	return capacity, nil
}

func getOnlineInstances(storagecluster common.Storage, nodes []cephclient.Node) []storagev1.Instance {
	instances := make([]storagev1.Instance, 0)
	for _, n := range nodes {
		name, err := cephclient.GetIDToHost(strconv.FormatInt(n.ID, 10))
		if err != nil {
			continue
		}
		stat := true
		if n.Total == 0 {
			stat = false
		}
		host, dev := osdSplit(storagecluster, name)
		if host == "" && dev == "" {
			continue
		}
		info := storagev1.Instance{
			Host: host,
			Dev:  dev,
			Stat: stat,
			Info: storagev1.Size{
				Total: strconv.FormatInt(n.Total*unit, 10),
				Used:  strconv.FormatInt(n.Used*unit, 10),
				Free:  strconv.FormatInt(n.Avail*unit, 10),
			},
		}
		instances = append(instances, info)
	}
	return instances
}

func getOfflineInstances(storagecluster common.Storage, onlineinstances []storagev1.Instance) []storagev1.Instance {
	instances := make([]storagev1.Instance, 0)
	online := make(map[string][]string)
	for _, i := range onlineinstances {
		if len(online[i.Host]) == 0 {
			online[i.Host] = make([]string, 0)
		}
		online[i.Host] = append(online[i.Host], i.Dev)
	}
	onlinecluster := common.MakeClusterCfg(online, "ceph")
	delcfg, _, _, _ := common.Diff(storagecluster, onlinecluster)
	for host, devs := range delcfg {
		if len(devs) == 0 {
			instance := storagev1.Instance{
				Host: host,
				Dev:  "",
				Stat: false,
			}
			instances = append(instances, instance)
		}
		for _, dev := range devs {
			instance := storagev1.Instance{
				Host: host,
				Dev:  dev,
				Stat: false,
			}
			instances = append(instances, instance)
		}
	}
	return instances
}

func osdSplit(storagecluster common.Storage, podname string) (string, string) {
	for _, h := range storagecluster.Spec.Hosts {
		for _, d := range h.BlockDevices {
			if strings.Contains(podname, h.NodeName) && strings.Contains(podname, d[5:]) {
				return h.NodeName, d
			}
		}
	}
	return "", ""
}
