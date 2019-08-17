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
		if storagecluster.Status.Phase == "Updating" || storagecluster.Status.Phase == "Creating" {
			continue
		}
		phase, message, capacity, err := getStatus(storagecluster)
		if err != nil {
			log.Warnf("[ceph-status-controller] Get ceph status failed. Err: %s", err.Error())
			continue
		}
		if err := common.UpdateStatus(cli, name, phase, message, capacity); err != nil {
			log.Warnf("[ceph-status-controller] Update storage cluster %s failed. Err: %s", name, err.Error())
		}
	}
}

func getStatus(storagecluster storagev1.Cluster) (string, string, storagev1.Capacity, error) {
	var phase, message string
	var capacity storagev1.Capacity

	phase, message, err := getPhaseAndMsg()
	if err != nil {
		return phase, message, capacity, err
	}
	capacity, err = getCapacity(storagecluster)
	if err != nil {
		return phase, message, capacity, err
	}
	return phase, message, capacity, nil
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

func getCapacity(storagecluster storagev1.Cluster) (storagev1.Capacity, error) {
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

func getOnlineInstances(storagecluster storagev1.Cluster, nodes []cephclient.Node) []storagev1.Instance {
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

func getOfflineInstances(storagecluster storagev1.Cluster, onlineinstances []storagev1.Instance) []storagev1.Instance {
	instances := make([]storagev1.Instance, 0)
	online := make(map[string][]string)
	for _, i := range onlineinstances {
		if len(online[i.Host]) == 0 {
			online[i.Host] = make([]string, 0)
		}
		online[i.Host] = append(online[i.Host], i.Dev)
	}
	onlinecluster := mapToInfos(online)
	delcfg, _, _, _ := common.Diff(storagecluster.Status.Config, onlinecluster)
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

func osdSplit(storagecluster storagev1.Cluster, podname string) (string, string) {
	for _, h := range storagecluster.Status.Config {
		for _, d := range h.BlockDevices {
			str1 := "-" + h.NodeName + "-"
			str2 := "-" + d.Name[5:] + "-"
			if strings.Contains(podname, str1) && strings.Contains(podname, str2) {
				return h.NodeName, d.Name
			}
		}
	}
	return "", ""
}

func mapToInfos(online map[string][]string) []storagev1.HostInfo {
	hosts := make([]storagev1.HostInfo, 0)
	for k, v := range online {
		devs := make([]storagev1.Dev, 0)
		for _, d := range v {
			dev := storagev1.Dev{
				Name: d,
			}
			devs = append(devs, dev)
		}
		host := storagev1.HostInfo{
			NodeName:     k,
			BlockDevices: devs,
		}
		hosts = append(hosts, host)
	}
	return hosts
}
