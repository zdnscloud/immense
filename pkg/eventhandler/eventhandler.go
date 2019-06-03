package eventhandler

import (
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/eventhandler/lvm"
)

func Create(cli client.Client, cluster *storagev1.Cluster) error {
	switch cluster.Spec.StorageType {
	case "lvm":
		return lvm.Create(cli, cluster)
	}
	return nil
}

func Delete(cli client.Client, cluster *storagev1.Cluster) error {
	switch cluster.Spec.StorageType {
	case "lvm":
		return lvm.Delete(cli, cluster)
	}
	return nil
}

func Update(cli client.Client, oldc *storagev1.Cluster, newc *storagev1.Cluster) error {
	/*
		switch oldc.Spec.StorageType {
		case "lvm":
			return lvm.Update(cli, oldc, newc)
		}*/
	return nil
}
