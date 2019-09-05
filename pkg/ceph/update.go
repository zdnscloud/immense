package ceph

import (
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/ceph/osd"
)

func doDelhost(cli client.Client, cluster storagev1.Cluster) error {
	if len(cluster.Spec.Hosts) == 0 {
		return nil
	}
	for _, host := range cluster.Status.Config {
		for _, d := range host.BlockDevices {
			dev := d[5:]
			if err := osd.Stop(cli, host.NodeName, dev); err != nil {
				return err
			}
		}
	}
	return nil
}

func doAddhost(cli client.Client, cluster storagev1.Cluster) error {
	if len(cluster.Spec.Hosts) == 0 {
		return nil
	}
	for _, host := range cluster.Status.Config {
		for _, d := range host.BlockDevices {
			dev := d[5:]
			if err := osd.Start(cli, host.NodeName, dev); err != nil {
				return err
			}
		}
	}
	return nil
}
