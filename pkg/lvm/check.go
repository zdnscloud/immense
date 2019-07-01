package lvm

import (
	"context"
	"github.com/zdnscloud/gok8s/client"
	storagev1 "k8s.io/api/storage/v1"
)

func CheckUsed(cli client.Client, node string) (bool, error) {
	attacher := "csi-lvmplugin"
	vas := storagev1.VolumeAttachmentList{}
	err := cli.List(context.TODO(), nil, &vas)
	if err != nil {
		return false, err
	}
	for _, v := range vas.Items {
		if v.Spec.Attacher == attacher && v.Spec.NodeName == node && v.Status.Attached {
			return true, nil
		}
	}
	return false, nil
}
