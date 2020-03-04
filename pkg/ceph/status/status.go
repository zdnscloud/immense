package status

import (
	"strconv"
	"strings"
	"time"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	cephclient "github.com/zdnscloud/immense/pkg/ceph/client"
	"github.com/zdnscloud/immense/pkg/common"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func Watch(cli client.Client, name string) {
	log.Debugf("[ceph-status-controller] Start")
	for {
		time.Sleep(60 * time.Second)
		storagecluster, err := common.GetStorageCluster(cli, name)
		if err != nil {
			if apierrors.IsNotFound(err) == false {
				log.Warnf("[ceph-status-controller] Get storage cluster %s failed. Err: %s", name, err.Error())
			}
			log.Debugf("[ceph-status-controller] Stop")
			return
		}
		if storagecluster.DeletionTimestamp != nil {
			log.Debugf("[ceph-status-controller] Stop")
			return
		}
		if storagecluster.Status.Phase == "Updating" || storagecluster.Status.Phase == "Creating" {
			continue
		}
		status, err := genStatus(storagecluster)
		if err != nil {
			log.Warnf("[ceph-status-controller] Get ceph status failed. Err: %s", err.Error())
			continue
		}
		if err := common.UpdateClusterStatus(cli, storagecluster.Name, status); err != nil {
			log.Warnf("[ceph-status-controller] Update storage cluster %s failed. Err: %s", name, err.Error())
		}
	}
}

func genStatus(storagecluster *storagev1.Cluster) (storagev1.ClusterStatus, error) {
	var status storagev1.ClusterStatus
	var err error
	status.Phase, status.Message, err = getPhaseAndMsg()
	if err != nil {
		return status, err
	}
	status.Capacity, err = getCapacity(storagecluster)
	if err != nil {
		return status, err
	}
	status.Config = storagecluster.Status.Config
	return status, nil
}

func getPhaseAndMsg() (storagev1.StatusPhase, string, error) {
	var (
		phase   storagev1.StatusPhase
		message string
	)
	message, err := cephclient.CheckHealth()
	if err != nil {
		return phase, message, err
	}
	if strings.Contains(message, "HEALTH_OK") {
		phase = storagev1.Running
	} else if strings.Contains(message, "HEALTH_WARN") {
		phase = storagev1.Warnning
	} else {
		phase = storagev1.Failed
	}
	return phase, message, nil
}

var unit = int64(1024)

func getCapacity(storagecluster *storagev1.Cluster) (storagev1.Capacity, error) {
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

func getOnlineInstances(storagecluster *storagev1.Cluster, nodes []cephclient.Node) []storagev1.Instance {
	instances := make([]storagev1.Instance, 0)
	for _, n := range nodes {
		name, err := cephclient.GetIDToHost(strconv.FormatInt(n.ID, 10))
		if err != nil {
			continue
		}
		stat := true
		if n.Total == 0 || n.Status != "up" {
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

func getOfflineInstances(storagecluster *storagev1.Cluster, onlineinstances []storagev1.Instance) []storagev1.Instance {
	instances := make([]storagev1.Instance, 0)
	online := make(map[string][]string)
	for _, i := range onlineinstances {
		if len(online[i.Host]) == 0 {
			online[i.Host] = make([]string, 0)
		}
		online[i.Host] = append(online[i.Host], i.Dev)
	}
	diff := getDiff(storagecluster.Status.Config, online)
	for _, info := range diff {
		if len(info.BlockDevices) == 0 {
			instance := storagev1.Instance{
				Host: info.NodeName,
				Dev:  "",
				Stat: false,
			}
			instances = append(instances, instance)
		}
		for _, dev := range info.BlockDevices {
			instance := storagev1.Instance{
				Host: info.NodeName,
				Dev:  dev,
				Stat: false,
			}
			instances = append(instances, instance)
		}
	}
	return instances
}

func osdSplit(storagecluster *storagev1.Cluster, podname string) (string, string) {
	for _, h := range storagecluster.Status.Config {
		for _, d := range h.BlockDevices {
			str1 := "-" + h.NodeName + "-"
			str2 := "-" + d[5:] + "-"
			if strings.Contains(podname, str1) && strings.Contains(podname, str2) {
				return h.NodeName, d
			}
		}
	}
	return "", ""
}

func getDiff(oldcfg []storagev1.HostInfo, online map[string][]string) []storagev1.HostInfo {
	delcfg := make([]storagev1.HostInfo, 0)
	for _, info := range oldcfg {
		_, ok := online[info.NodeName]
		if ok {
			continue
		}
		delcfg = append(delcfg, info)
	}
	return delcfg
}
