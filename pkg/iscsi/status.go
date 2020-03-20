package iscsi

import (
	"fmt"
	"strconv"
	"time"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/common"
	pb "github.com/zdnscloud/lvmd/proto"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func StatusControl(cli client.Client, name string) {
	ctrlName := fmt.Sprintf("[%s-iscsi-status-controller]", name)
	log.Debugf("%s Start", ctrlName)
	for {
		time.Sleep(60 * time.Second)
		iscsi, err := getIscsi(cli, name)
		if err != nil {
			if apierrors.IsNotFound(err) == false {
				log.Warnf("%s Get storage failed. Err: %s", ctrlName, err.Error())
				continue
			}
			log.Debugf("%s Stop", ctrlName)
			return
		}
		if iscsi.DeletionTimestamp != nil {
			log.Debugf("%s Stop", ctrlName)
			return
		}

		if iscsi.Status.Phase == storagev1.Updating || iscsi.Status.Phase == storagev1.Creating || iscsi.Status.Phase == storagev1.Failed {
			continue
		}
		status := genStatus(cli, iscsi, ctrlName)
		if err := updateStatus(cli, name, status); err != nil {
			log.Warnf("%s Update storage failed. Err: %s", ctrlName, err.Error())
			continue
		}
	}
}

func genStatus(cli client.Client, iscsi *storagev1.Iscsi, ctrlName string) storagev1.IscsiStatus {
	var status storagev1.IscsiStatus
	instances := make([]storagev1.Instance, 0)
	for _, host := range iscsi.Spec.Initiators {
		var instance storagev1.Instance
		instance.Host = host
		lvmdcli, err := common.CreateLvmdClientForPod(cli, host, common.StorageNamespace, fmt.Sprintf("%s-%s", iscsi.Name, IscsiLvmdDsSuffix))
		if err != nil {
			status.Phase = storagev1.Warnning
			status.Message = status.Message + host + ":" + err.Error() + "\n"
			instance.Stat = false
			instances = append(instances, instance)
			log.Warnf("%s Connect to %s lvmd faield. Err: %s", ctrlName, host, err.Error())
			continue
		}
		defer lvmdcli.Close()
		vgsreq := pb.ListVGRequest{}
		vgsout, err := lvmdcli.ListVG(ctx, &vgsreq)
		if err != nil {
			status.Phase = storagev1.Warnning
			status.Message = status.Message + host + ":" + err.Error() + "\n"
			instance.Stat = false
			instances = append(instances, instance)
			log.Warnf("%s List volume group faield for host %s . Err: %s", ctrlName, host, err.Error())
			continue
		}
		if len(status.Phase) == 0 {
			status.Phase = storagev1.Running
		}
		for _, v := range vgsout.VolumeGroups {
			if v.Name != fmt.Sprintf("%s-%s", iscsi.Name, VolumeGroupSuffix) {
				continue
			}
			instance.Stat = true
			instance.Info = storagev1.Size{
				Total: string(strconv.Itoa(int(v.Size))),
				Used:  string(strconv.Itoa(int(v.Size - v.FreeSize))),
				Free:  string(strconv.Itoa(int(v.FreeSize))),
			}
			instances = append(instances, instance)
		}
	}
	if len(instances) == 0 {
		status.Phase = storagev1.Failed
	}
	status.Capacity.Instances = instances
	if len(instances) > 0 {
		status.Capacity.Total = storagev1.Size{
			Total: instances[0].Info.Total,
			Used:  instances[0].Info.Used,
			Free:  instances[0].Info.Free,
		}
	}
	return status
}
