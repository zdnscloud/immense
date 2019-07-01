package ceph

import (
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/immense/pkg/ceph/osd"
)

func doDelhost(cli client.Client, cfg map[string][]string) error {
	if len(cfg) == 0 {
		return nil
	}
	log.Debugf("Delete device for storage, Cfg: %s", cfg)
	for host, devs := range cfg {
		for _, d := range devs {
			dev := d[5:]
			if err := osd.Remove(cli, host, dev); err != nil {
				return err
			}
		}
	}
	return nil
}

func doAddhost(cli client.Client, cfg map[string][]string) error {
	if len(cfg) == 0 {
		return nil
	}
	log.Debugf("Add device for storage, Cfg: %s", cfg)
	for host, devs := range cfg {
		for _, d := range devs {
			dev := d[5:]
			if err := osd.Start(cli, host, dev); err != nil {
				return err
			}
		}
	}
	return nil
}

func doChangeAdd(cli client.Client, cfg map[string][]string) error {
	if len(cfg) == 0 {
		return nil
	}
	return doAddhost(cli, cfg)
}

func doChangeDel(cli client.Client, cfg map[string][]string) error {
	if len(cfg) == 0 {
		return nil
	}
	return doDelhost(cli, cfg)
}
