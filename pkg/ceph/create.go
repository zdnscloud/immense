package ceph

import (
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
	"github.com/zdnscloud/immense/pkg/ceph/status"
	"github.com/zdnscloud/immense/pkg/ceph/util"
	"github.com/zdnscloud/immense/pkg/common"
	"strings"
)

func create(cli client.Client, cluster *storagev1.Cluster) error {
	/*
		networks, err := util.GetCIDRs(cli, cluster)
		if err != nil {
			return err
		}*/
	networks := "10.42.0.0/16"
	uuid, adminkey, monkey, err := initconf()
	if err != nil {
		return err
	}
	copiers := getReplication(cluster)
	s := storagev1.ClusterStatus{
		Phase: "Creating"}
	if err := common.UpdateStatus(cli, cluster.Name, s); err != nil {
		log.Warnf("Update storage cluster %s status failed. Err: %s", cluster.Name, err.Error())
	}
	if err := config.Start(cli, uuid, networks, adminkey, monkey, copiers); err != nil {
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
	go osd.Watch()
	go mon.Watch(cli)
	go status.Watch(cli, cluster.Name)
	s = storagev1.ClusterStatus{
		Phase: "Running"}
	if err := common.UpdateStatus(cli, cluster.Name, s); err != nil {
		log.Warnf("Update storage cluster %s status failed. Err: %s", cluster.Name, err.Error())
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

func getReplication(cluster *storagev1.Cluster) int {
	var num, Replication int
	for _, host := range cluster.Spec.Hosts {
		num += len(host.BlockDevices)
	}
	if num > 2 {
		Replication = global.PoolDefaultSize
	} else {
		Replication = 1
	}
	return Replication
}
