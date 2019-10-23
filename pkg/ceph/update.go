package ceph

import (
	"errors"
	"github.com/zdnscloud/cement/errgroup"
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"strings"

	"github.com/zdnscloud/immense/pkg/ceph/osd"
	"github.com/zdnscloud/immense/pkg/ceph/prepare"
	"github.com/zdnscloud/immense/pkg/ceph/util"
)

func doDelhost(cli client.Client, cluster storagev1.Cluster) error {
	if len(cluster.Spec.Hosts) == 0 {
		return nil
	}
	if err := osd.Stop(cli, cluster.Status.Config); err != nil {
		return err
	}
	_, err := errgroup.Batch(
		cluster.Spec.Hosts,
		func(o interface{}) (interface{}, error) {
			host := o.(string)
			return nil, prepare.Delete(cli, host)
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func doAddhost(cli client.Client, cluster storagev1.Cluster) error {
	if len(cluster.Spec.Hosts) == 0 {
		return nil
	}
	uuid, err := util.GetCephUUID(cli)
	if err != nil {
		return err
	}
	if uuid == "" {
		return errors.New("can not get storage cluster uuid")
	}
	monsvc, err := util.GetMonSvcMap(cli)
	if err != nil {
		return err
	}
	_, err = errgroup.Batch(
		cluster.Spec.Hosts,
		func(o interface{}) (interface{}, error) {
			host := o.(string)
			devs := util.GetDevsForHost(cluster, host)
			return nil, prepare.Do(cli, host, devs)
		},
	)
	if err != nil {
		return err
	}
	_, err = errgroup.Batch(
		util.ToSlice(cluster),
		func(o interface{}) (interface{}, error) {
			host := strings.Split(o.(string), ":")[0]
			dev := strings.Split(o.(string), ":")[1][5:]
			return nil, osd.Start(cli, uuid, host, dev, monsvc)
		},
	)
	if err != nil {
		return err
	}

	return nil
}
