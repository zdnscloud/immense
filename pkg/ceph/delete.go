package ceph

import (
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/zdnscloud/cement/errgroup"
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/ceph/config"
	"github.com/zdnscloud/immense/pkg/ceph/fscsi"
	"github.com/zdnscloud/immense/pkg/ceph/mds"
	"github.com/zdnscloud/immense/pkg/ceph/mgr"
	"github.com/zdnscloud/immense/pkg/ceph/mon"
	"github.com/zdnscloud/immense/pkg/ceph/osd"
	"github.com/zdnscloud/immense/pkg/ceph/prepare"
	"github.com/zdnscloud/immense/pkg/ceph/util"
)

func delete(cli client.Client, cluster storagev1.Cluster) error {
	var uuid, adminkey, monkey string
	var copies int
	if err := fscsi.Stop(cli, uuid, cluster.Name); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	if err := mds.Stop(cli); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	_, err := errgroup.Batch(
		util.ToSlice(cluster),
		func(o interface{}) (interface{}, error) {
			host := strings.Split(o.(string), ":")[0]
			dev := strings.Split(o.(string), ":")[1][5:]
			return nil, osd.Remove(cli, host, dev)
		},
	)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	_, err = errgroup.Batch(
		cluster.Spec.Hosts,
		func(o interface{}) (interface{}, error) {
			host := o.(string)
			return nil, prepare.Delete(cli, host)
		},
	)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	if err := mgr.Stop(cli); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	if err := mon.Stop(cli); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	if err := config.Stop(cli, uuid, adminkey, monkey, copies); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	if err := util.RemoveConf(cli); err != nil {
		return err
	}
	return nil
}
