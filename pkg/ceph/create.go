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

func create(cli client.Client, cluster *storagev1.Cluster) error {
	networks, err := util.GetCIDRs(cli, cluster)
	if err != nil {
		return err
	}
	networks = "10.42.0.0/16"
	uuid, adminkey, monkey, err := initconf()
	if err != nil {
		return err
	}
	if err := config.Start(cli, uuid, networks, adminkey, monkey); err != nil {
		return err
	}
	if err := cephclient.SaveConf(cli); err != nil {
		return err
	}
	if err := mon.Start(cli, networks); err != nil {
		return err
	}
	if err := mgr.Start(cli); err != nil {
		return err
	}
	_, err = errgroup.Batch(
		util.ToSlice(cluster),
		func(o interface{}) (interface{}, error) {
			host := strings.Split(o.(string), ":")[0]
			dev := strings.Split(o.(string), ":")[1][5:]
			return nil, osd.Start(cli, host, dev)
		},
	)
	if err != nil {
		return err
	}
	if err := mds.Start(cli); err != nil {
		return err
	}
	if err := fscsi.Start(cli); err != nil {
		return err
	}
	return nil
}

func initconf() (string, string, string, error) {
	var uuid, adminkey, monkey string
	uuid, err := util.ExecCMDWithOutput("/usr/bin/uuidgen", []string{})
	if err != nil {
		return uuid, adminkey, monkey, err
	}
	adminkey, err = util.ExecCMDWithOutput("/usr/bin/python", []string{"/ceph-key.py"})
	if err != nil {
		return uuid, adminkey, monkey, err
	}
	monkey, err = util.ExecCMDWithOutput("/usr/bin/python", []string{"/ceph-key.py"})
	if err != nil {
		return uuid, adminkey, monkey, err
	}
	return uuid, adminkey, monkey, nil
}