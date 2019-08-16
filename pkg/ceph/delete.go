package ceph

import (
	"github.com/zdnscloud/cement/errgroup"
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	cephclient "github.com/zdnscloud/immense/pkg/ceph/client"
	"github.com/zdnscloud/immense/pkg/ceph/config"
	"github.com/zdnscloud/immense/pkg/ceph/fscsi"
	"github.com/zdnscloud/immense/pkg/ceph/mds"
	"github.com/zdnscloud/immense/pkg/ceph/mgr"
	"github.com/zdnscloud/immense/pkg/ceph/mon"
	"github.com/zdnscloud/immense/pkg/ceph/osd"
	"github.com/zdnscloud/immense/pkg/ceph/util"
	"strings"
)

func delete(cli client.Client, cluster storagev1.Cluster) error {
	var networks, uuid, adminkey, monkey string
	var copies int
	if err := fscsi.Stop(cli); err != nil {
		return err
	}
	if err := mds.Stop(cli); err != nil {
		return err
	}
	_, err := errgroup.Batch(
		util.ToSlice(cluster),
		func(o interface{}) (interface{}, error) {
			host := strings.Split(o.(string), ":")[0]
			dev := strings.Split(o.(string), ":")[1][5:]
			return nil, osd.Stop(cli, host, dev)
		},
	)
	if err != nil {
		return err
	}
	if err := mgr.Stop(cli); err != nil {
		return err
	}
	if err := mon.Stop(cli, networks); err != nil {
		return err
	}
	if err := config.Stop(cli, uuid, networks, adminkey, monkey, copies); err != nil {
		return err
	}
	if err := cephclient.RemoveConf(cli); err != nil {
		return err
	}
	return nil
}
