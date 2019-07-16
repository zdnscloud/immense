package ceph

import (
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/immense/pkg/ceph/osd"
	"github.com/zdnscloud/immense/pkg/common"
)

func doDelhost(cli client.Client, cluster common.Storage) error {
	if len(cluster.Spec.Hosts) == 0 {
		return nil
	}
	log.Debugf("Delete device for storage, Cfg: %s", cluster)
	for _, host := range cluster.Spec.Hosts {
		for _, d := range host.BlockDevices {
			dev := d[5:]
			if err := osd.Stop(cli, host.NodeName, dev); err != nil {
				return err
			}
		}
	}
	return nil
}

func doAddhost(cli client.Client, cluster common.Storage) error {
	if len(cluster.Spec.Hosts) == 0 {
		return nil
	}
	log.Debugf("Add device for storage, Cfg: %s", cluster)
	for _, host := range cluster.Spec.Hosts {
		for _, d := range host.BlockDevices {
			dev := d[5:]
			if err := osd.Start(cli, host.NodeName, dev); err != nil {
				return err
			}
		}
	}
	return nil
}
