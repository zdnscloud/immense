package lvm

import (
	"context"
	"fmt"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/common"
	"github.com/zdnscloud/immense/pkg/lvm/util"
	pb "github.com/zdnscloud/lvmd/proto"
	"strconv"
	"strings"
	"time"
)

func StatusControl(cli client.Client, name string) {
	log.Debugf("[lvm-status-controller] Start")
	for {
		time.Sleep(60 * time.Second)
		storagecluster, err := common.GetStorage(cli, name)
		if err != nil {
			log.Warnf("[lvm-status-controller] Get storage %s config with blocks failed. Err: %s", name, err.Error())
			log.Debugf("[lvm-status-controller] Stop")
			return
		}
		if storagecluster.Status.Phase == "Updating" || storagecluster.Status.Phase == "Creating" {
			continue
		}
		phase, message, capacity := getStatus(cli, storagecluster)
		if err := common.UpdateStatus(cli, name, phase, message, capacity); err != nil {
			log.Warnf("[lvm-status-controller] Update storage cluster %s failed. Err: %s", name, err.Error())
			continue
		}
	}
}

func getStatus(cli client.Client, storagecluster storagev1.Cluster) (string, string, storagev1.Capacity) {
	ctx := context.TODO()
	var state, message string
	var capacity storagev1.Capacity
	instances := make([]storagev1.Instance, 0)
	var total, used, free uint64
	for _, host := range storagecluster.Status.Config {
		var instance storagev1.Instance
		instance.Host = host.NodeName
		devs := make([]string, 0)
		for _, d := range host.BlockDevices {
			devs = append(devs, d)
		}
		instance.Dev = strings.Replace(strings.Trim(fmt.Sprint(devs), "[]"), " ", ",", -1)
		if len(instance.Dev) == 0 {
			state = "Warnning"
			message = message + host.NodeName + ":" + "No block devices can be used\n"
			instance.Stat = false
			instances = append(instances, instance)
			log.Warnf("[lvm-status-controller] Hosts %s have no block devices can used", host.NodeName)
			continue
		}
		lvmdcli, err := util.CreateLvmdClient(ctx, cli, host.NodeName)
		if err != nil {
			state = "Warnning"
			message = message + host.NodeName + ":" + err.Error() + "\n"
			instance.Stat = false
			instances = append(instances, instance)
			log.Warnf("[lvm-status-controller] Connect to %s lvmd faield. Err: %s", host.NodeName, err.Error())
			continue
		}
		//instance.Stat = true
		vgsreq := pb.ListVGRequest{}
		vgsout, err := lvmdcli.ListVG(ctx, &vgsreq)
		if err != nil {
			state = "Warnning"
			message = message + host.NodeName + ":" + err.Error() + "\n"
			instance.Stat = false
			instances = append(instances, instance)
			log.Warnf("[lvm-status-controller] List volume group faield for host %s. Err: %s", host.NodeName, err.Error())
			continue
		}
		if len(state) == 0 {
			state = "Running"
		}
		for _, v := range vgsout.VolumeGroups {
			if v.Name != "k8s" {
				continue
			}
			instance.Stat = true
			instance.Info = storagev1.Size{
				Total: string(strconv.Itoa(int(v.Size))),
				Used:  string(strconv.Itoa(int(v.Size - v.FreeSize))),
				Free:  string(strconv.Itoa(int(v.FreeSize))),
			}
			instances = append(instances, instance)
			total += v.Size
			used += v.Size - v.FreeSize
			free += v.FreeSize
		}
	}
	if len(instances) == 0 {
		state = "Error"
	}
	capacity.Instances = instances
	capacity.Total = storagev1.Size{
		Total: string(strconv.Itoa(int(total))),
		Used:  string(strconv.Itoa(int(used))),
		Free:  string(strconv.Itoa(int(free))),
	}
	return state, message, capacity
}
