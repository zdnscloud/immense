package lvm

import (
	"fmt"
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/eventhandler/common"
)

func Update(cli client.Client, oldcfg, newcfg *storagev1.Cluster) error {
	delcfg, addcfg, changetodel, changetoadd := common.Diff(oldcfg, newcfg)
	fmt.Println(delcfg)
	fmt.Println(addcfg)
	fmt.Println(changetodel)
	fmt.Println(changetoadd)
	if err := doDelhost(cli, delcfg); err != nil {
		return err
	}
	if err := doAddhost(cli, addcfg); err != nil {
		return err
	}
	return nil
}

func doDelhost(cli client.Client, cfg map[string][]string) error {
	cluster := makeClusterCfg(cfg)
	if err := common.DeleteNodeAnnotationsAndLabels(cli, cluster, NodeLabelValue); err != nil {
		return err
	}
	return nil
}

func doAddhost(cli client.Client, cfg map[string][]string) error {
	cluster := makeClusterCfg(cfg)
	if err := common.CreateNodeAnnotationsAndLabels(cli, cluster, NodeLabelValue); err != nil {
		return err
	}
	if err := initBlocks(cli, cluster); err != nil {
		return err
	}
	return nil
}

func doChangeAdd(cli client.Client, cfg map[string][]string) error {
	cluster := makeClusterCfg(cfg)
	if err := initBlocks(cli, cluster); err != nil {
		return err
	}
	return nil
}

func makeClusterCfg(cfg map[string][]string) *storagev1.Cluster {
	hosts := make([]storagev1.HostSpec, 0)

	for k, v := range cfg {
		host := storagev1.HostSpec{
			NodeName:     k,
			BlockDevices: v,
		}
		hosts = append(hosts, host)
	}
	return &storagev1.Cluster{
		Spec: storagev1.ClusterSpec{
			StorageType: "lvm",
			Hosts:       hosts,
		},
	}
}
