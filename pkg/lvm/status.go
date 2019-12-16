package lvm

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/common"
	pb "github.com/zdnscloud/lvmd/proto"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

var ctx = context.TODO()

func StatusControl(cli client.Client, name string) {
	log.Debugf("[lvm-status-controller] Start")
	for {
		time.Sleep(60 * time.Second)
		storagecluster, err := common.GetStorage(cli, name)
		if err != nil {
			if apierrors.IsNotFound(err) == false {
				log.Warnf("[lvm-status-controller] Get storage cluster %s failed. Err: %s", name, err.Error())
			}
			log.Debugf("[lvm-status-controller] Stop")
			return
		}
		if storagecluster.DeletionTimestamp != nil {
			log.Debugf("[lvm-status-controller] Stop")
			return
		}

		if storagecluster.Status.Phase == storagev1.Updating || storagecluster.Status.Phase == storagev1.Creating {
			continue
		}
		storagecluster.Status = genStatus(cli, storagecluster)
		if err := cli.Update(ctx, &storagecluster); err != nil {
			log.Warnf("[lvm-status-controller] Update storage cluster %s failed. Err: %s", name, err.Error())
			continue
		}
	}
}

func genStatus(cli client.Client, storagecluster storagev1.Cluster) storagev1.ClusterStatus {
	var status storagev1.ClusterStatus
	status.Config = storagecluster.Status.Config
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
			status.Phase = storagev1.Warnning
			status.Message = status.Message + host.NodeName + ":" + "No block devices can be used\n"
			instance.Stat = false
			instances = append(instances, instance)
			log.Warnf("[lvm-status-controller] Hosts %s have no block devices can used", host.NodeName)
			continue
		}
		lvmdcli, err := CreateLvmdClient(ctx, cli, host.NodeName)
		if err != nil {
			status.Phase = storagev1.Warnning
			status.Message = status.Message + host.NodeName + ":" + err.Error() + "\n"
			instance.Stat = false
			instances = append(instances, instance)
			log.Warnf("[lvm-status-controller] Connect to %s lvmd faield. Err: %s", host.NodeName, err.Error())
			continue
		}
		defer lvmdcli.Close()
		//instance.Stat = true
		vgsreq := pb.ListVGRequest{}
		vgsout, err := lvmdcli.ListVG(ctx, &vgsreq)
		if err != nil {
			status.Phase = storagev1.Warnning
			status.Message = status.Message + host.NodeName + ":" + err.Error() + "\n"
			instance.Stat = false
			instances = append(instances, instance)
			log.Warnf("[lvm-status-controller] List volume group faield for host %s. Err: %s", host.NodeName, err.Error())
			continue
		}
		if len(status.Phase) == 0 {
			status.Phase = storagev1.Running
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
		status.Phase = storagev1.Failed
	}
	status.Capacity.Instances = instances
	status.Capacity.Total = storagev1.Size{
		Total: string(strconv.Itoa(int(total))),
		Used:  string(strconv.Itoa(int(used))),
		Free:  string(strconv.Itoa(int(free))),
	}
	return status
}
