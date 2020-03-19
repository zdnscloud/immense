package nfs

import (
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	k8stypes "k8s.io/apimachinery/pkg/types"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/common"
)

func getNfs(cli client.Client, name string) (*storagev1.Nfs, error) {
	nfs := &storagev1.Nfs{}
	if err := cli.Get(ctx, k8stypes.NamespacedName{"", name}, nfs); err != nil {
		return nil, err
	}
	return nfs, nil
}

func AddFinalizer(cli client.Client, name, finalizer string) error {
	nfs, err := getNfs(cli, name)
	if err != nil {
		return err
	}
	helper.AddFinalizer(nfs, finalizer)
	log.Debugf("Add finalizer %s for storage nfs %s", finalizer, name)
	return cli.Update(ctx, nfs)
}

func RemoveFinalizer(cli client.Client, name, finalizer string) error {
	nfs, err := getNfs(cli, name)
	if err != nil {
		return err
	}
	helper.RemoveFinalizer(nfs, finalizer)
	log.Debugf("Delete finalizer %s for storage nfs %s", finalizer, name)
	return cli.Update(ctx, nfs)
}

func UpdateStatusPhase(cli client.Client, name string, phase storagev1.StatusPhase) {
	nfs, err := getNfs(cli, name)
	if err != nil {
		if apierrors.IsNotFound(err) == true {
			return
		}
		log.Warnf("Update storage iscsi %s status failed. Err: %s", name, err.Error())
		return
	}
	nfs.Status.Phase = phase
	if err := cli.Update(ctx, nfs); err != nil {
		if apierrors.IsNotFound(err) == true {
			return
		}
		log.Warnf("Update storage iscsi %s status failed. Err: %s", name, err.Error())
		return
	}
	return
}

func updateSize(cli client.Client, name string, size *storagev1.Size) error {
	nfs, err := getNfs(cli, name)
	if err != nil {
		return err
	}
	nfs.Status.Capacity.Total = *size
	return cli.Update(ctx, nfs)
}

func IsPvLastOne(cli client.Client, driver string) (bool, error) {
	pvs := corev1.PersistentVolumeList{}
	if err := cli.List(ctx, nil, &pvs); err != nil {
		return false, err
	}
	for _, pv := range pvs.Items {
		if _driver, ok := pv.Annotations[common.PvProvisionerKey]; ok && _driver == driver {
			return false, nil
		}
	}
	return true, nil
}
