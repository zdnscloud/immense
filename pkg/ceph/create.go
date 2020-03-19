package ceph

import (
	"strings"

	"github.com/zdnscloud/cement/errgroup"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	cephclient "github.com/zdnscloud/immense/pkg/ceph/client"
	"github.com/zdnscloud/immense/pkg/ceph/config"
	"github.com/zdnscloud/immense/pkg/ceph/fscsi"
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/ceph/mds"
	"github.com/zdnscloud/immense/pkg/ceph/mgr"
	"github.com/zdnscloud/immense/pkg/ceph/mon"
	"github.com/zdnscloud/immense/pkg/ceph/osd"
	"github.com/zdnscloud/immense/pkg/ceph/prepare"
	"github.com/zdnscloud/immense/pkg/ceph/status"
	"github.com/zdnscloud/immense/pkg/ceph/util"
)

func create(cli client.Client, cluster storagev1.Cluster) error {
	uuid := string(cluster.UID)
	adminkey, monkey, err := genkey()
	if err != nil {
		return err
	}
	copiers, pgnum := getReplicationAndPgNum(cluster.Status.Config)
	log.Debugf("Based on block device number, the number of replication is %d,pg_num is %d", copiers, pgnum)

	if err := config.Start(cli, uuid, adminkey, monkey, copiers); err != nil {
		return err
	}
	if err := util.SaveConf(cli); err != nil {
		return err
	}
	monsvc, err := util.GetMonSvcMap(cli)
	if err != nil {
		return err
	}
	if err := mon.Start(cli, uuid, monsvc); err != nil {
		return err
	}
	if err := mgr.Start(cli); err != nil {
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
	if err := mds.Start(cli, uuid, monsvc, copiers, pgnum); err != nil {
		return err
	}
	if err := cephclient.EnableDashboard(); err != nil {
		return err
	}
	if err := fscsi.Start(cli, uuid, cluster.Name, monsvc); err != nil {
		return err
	}
	go status.Watch(cli, cluster.Name)
	return nil
}

func genkey() (string, string, error) {
	var adminkey, monkey string
	adminkey, err := util.ExecCMDWithOutput("/usr/bin/python", []string{"/ceph-key.py"})
	if err != nil {
		return adminkey, monkey, err
	}
	monkey, err = util.ExecCMDWithOutput("/usr/bin/python", []string{"/ceph-key.py"})
	if err != nil {
		return adminkey, monkey, err
	}
	return adminkey, monkey, nil
}

func getReplicationAndPgNum(hosts []storagev1.HostInfo) (int, int) {
	var num, Replication, PgNum int
	for _, host := range hosts {
		num += len(host.BlockDevices)
	}
	if num > 1 {
		Replication = global.PoolDefaultSize
	} else {
		Replication = 1
	}
	PgNum = global.TargetPgPerOsd * num / 2 / Replication
	return Replication, PgNum
}
