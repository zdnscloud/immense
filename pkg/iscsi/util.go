package iscsi

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	k8stypes "k8s.io/apimachinery/pkg/types"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
)

func getIscsi(cli client.Client, name string) (*storagev1.Iscsi, error) {
	iscsi := &storagev1.Iscsi{}
	if err := cli.Get(ctx, k8stypes.NamespacedName{"", name}, iscsi); err != nil {
		return nil, err
	}
	return iscsi, nil
}

func AddFinalizer(cli client.Client, name, finalizer string) error {
	iscsi, err := getIscsi(cli, name)
	if err != nil {
		return err
	}
	helper.AddFinalizer(iscsi, finalizer)
	log.Debugf("Add finalizer %s for storage iscsi %s", finalizer, name)
	return cli.Update(ctx, iscsi)
}

func RemoveFinalizer(cli client.Client, name, finalizer string) error {
	iscsi, err := getIscsi(cli, name)
	if err != nil {
		return err
	}
	helper.RemoveFinalizer(iscsi, finalizer)
	log.Debugf("Delete finalizer %s for storage iscsi %s", finalizer, name)
	return cli.Update(ctx, iscsi)
}

func UpdateStatusPhase(cli client.Client, name string, phase storagev1.StatusPhase) {
	iscsi, err := getIscsi(cli, name)
	if err != nil {
		if apierrors.IsNotFound(err) == true {
			return
		}
		log.Warnf("Update storage iscsi %s status failed. Err: %s", name, err.Error())
		return
	}
	iscsi.Status.Phase = phase
	if err := cli.Update(ctx, iscsi); err != nil {
		if apierrors.IsNotFound(err) == true {
			return
		}
		log.Warnf("Update storage iscsi %s status failed. Err: %s", name, err.Error())
		return
	}
	return
}
