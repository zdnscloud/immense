package iscsi

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/common"
)

func update(cli client.Client, oldConf, newConf *storagev1.Iscsi) error {
	if !reflect.DeepEqual(oldConf.Spec.Targets, newConf.Spec.Targets) ||
		oldConf.Spec.Port != newConf.Spec.Port ||
		oldConf.Spec.Chap != newConf.Spec.Chap {
		if err := redeploy(cli, oldConf, newConf); err != nil {
			return err
		}
	} else if !reflect.DeepEqual(oldConf.Spec.Initiators, newConf.Spec.Initiators) {
		if err := updateInitiators(cli, oldConf, newConf); err != nil {
			return err
		}
	} else {
		return nil
	}
	return nil
}

func redeploy(cli client.Client, oldConf, newConf *storagev1.Iscsi) error {
	if err := delete(cli, oldConf); err != nil {
		return err
	}
	if err := create(cli, newConf); err != nil {
		return err
	}
	return nil
}

func updateInitiators(cli client.Client, oldConf, newConf *storagev1.Iscsi) error {
	dels, adds := common.HostsDiff(oldConf.Spec.Initiators, newConf.Spec.Initiators)
	if err := iscsiLoginAll(cli, newConf, adds); err != nil {
		UpdateStatusPhase(cli, newConf.Name, storagev1.Failed)
		return err
	}
	if err := iscsiLogoutAll(cli, oldConf, dels); err != nil {
		UpdateStatusPhase(cli, oldConf.Name, storagev1.Failed)
		return err
	}
	if err := common.CreateNodeAnnotationsAndLabels(cli, fmt.Sprintf("%s-%s", IscsiInstanceLabelKeyPrefix, newConf.Name), IscsiInstanceLabelValue, adds); err != nil {
		UpdateStatusPhase(cli, newConf.Name, storagev1.Failed)
		return err
	}
	if err := common.DeleteNodeAnnotationsAndLabels(cli, fmt.Sprintf("%s-%s", IscsiInstanceLabelKeyPrefix, newConf.Name), IscsiInstanceLabelValue, dels); err != nil {
		UpdateStatusPhase(cli, newConf.Name, storagev1.Failed)
		return err
	}
	time.Sleep(30 * time.Second)
	if err := common.WaitReady(common.DaemonSetObj(), cli, common.StorageNamespace, fmt.Sprintf("%s-%s", newConf.Name, IscsiLvmdDsSuffix)); err != nil {
		return err
	}
	if err := common.WaitReady(common.DaemonSetObj(), cli, common.StorageNamespace, fmt.Sprintf("%s-%s", newConf.Name, IscsiCSIDsSuffix)); err != nil {
		return err
	}
	if !checkVolumeGroup(cli, newConf) {
		UpdateStatusPhase(cli, newConf.Name, storagev1.Failed)
		return errors.New("can not get volumegroup from initiators")
	}
	return nil
}
