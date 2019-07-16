package common

import (
	"context"
	"encoding/json"
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"net/http"
	"strings"
)

func AssembleCreateConfig(cli client.Client, cluster *storagev1.Cluster) (Storage, error) {
	hosts := make([]Host, 0)
	for _, h := range cluster.Spec.Hosts {
		var host Host
		host.NodeName = h
		devs, err := GetBlocksFromAnnotation(cli, h)
		if err != nil {
			return Storage{}, err
		}
		if len(devs) > 0 {
			host.BlockDevices = devs
			hosts = append(hosts, host)
			continue
		}
		devs, err = GetBlocksFromClusterAgent(cli, h)
		if err != nil {
			return Storage{}, err
		}
		host.BlockDevices = devs
		hosts = append(hosts, host)
	}
	return Storage{
		Name: cluster.Name,
		Spec: StorageSpec{
			StorageType: cluster.Spec.StorageType,
			Hosts:       hosts,
		},
	}, nil
}

func AssembleDeleteConfig(cli client.Client, cluster *storagev1.Cluster) (Storage, error) {
	hosts := make([]Host, 0)
	for _, h := range cluster.Spec.Hosts {
		var host Host
		host.NodeName = h
		devs, err := GetBlocksFromAnnotation(cli, h)
		if err != nil {
			return Storage{}, err
		}
		host.BlockDevices = devs
		hosts = append(hosts, host)
	}
	return Storage{
		Name: cluster.Name,
		Spec: StorageSpec{
			StorageType: cluster.Spec.StorageType,
			Hosts:       hosts,
		},
	}, nil
}

func AssembleUpdateConfig(cli client.Client, oldc, newc *storagev1.Cluster) (Storage, Storage, error) {
	del, add := HostsDiff(oldc.Spec.Hosts, newc.Spec.Hosts)
	delc := &storagev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: newc.Name,
		},
		Spec: storagev1.ClusterSpec{
			StorageType: newc.Spec.StorageType,
			Hosts:       del,
		},
	}
	addc := &storagev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: newc.Name,
		},
		Spec: storagev1.ClusterSpec{
			StorageType: newc.Spec.StorageType,
			Hosts:       add,
		},
	}
	dels, err := AssembleDeleteConfig(cli, delc)
	if err != nil {
		return Storage{}, Storage{}, err
	}
	adds, err := AssembleCreateConfig(cli, addc)
	if err != nil {
		return Storage{}, Storage{}, err
	}
	return dels, adds, nil
}

func GetBlocksFromAnnotation(cli client.Client, name string) ([]string, error) {
	node := corev1.Node{}
	if err := cli.Get(ctx, k8stypes.NamespacedName{"", name}, &node); err != nil {
		return []string{}, err
	}
	blocks, ok := node.Annotations[StorageBlocksAnnotations]
	if ok {
		return strings.Split(blocks, ","), nil
	}
	return []string{}, nil
}

func GetBlocksFromClusterAgent(cli client.Client, name string) ([]string, error) {
	res := make([]string, 0)
	service := corev1.Service{}
	err := cli.Get(context.TODO(), k8stypes.NamespacedName{StorageNamespace, "cluster-agent"}, &service)
	if err != nil {
		return res, err
	}
	url := "/apis/agent.zcloud.cn/v1/blockinfos"
	newurl := "http://" + service.Spec.ClusterIP + url
	req, err := http.NewRequest("GET", newurl, nil)
	if err != nil {
		return res, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return res, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var info Data
	json.Unmarshal(body, &info)
	for _, v := range info.Data {
		for _, h := range v.Hosts {
			if h.NodeName != name {
				continue
			}
			for _, dev := range h.BlockDevices {
				if dev.Parted || dev.Filesystem || dev.Mount {
					continue
				}
				res = append(res, dev.Name)
			}
		}
	}
	return res, nil
}

func GetStorage(cli client.Client, name string) (Storage, error) {
	storagecluster := storagev1.Cluster{}
	err := cli.Get(context.TODO(), k8stypes.NamespacedName{StorageNamespace, name}, &storagecluster)
	if err != nil {
		return Storage{}, err
	}
	hosts := make([]Host, 0)
	for _, host := range storagecluster.Spec.Hosts {
		devs, _ := GetBlocksFromAnnotation(cli, host)
		host := Host{
			NodeName:     host,
			BlockDevices: devs,
		}
		hosts = append(hosts, host)
	}
	return Storage{
		Name: name,
		Spec: StorageSpec{
			StorageType: storagecluster.Spec.StorageType,
			Hosts:       hosts,
		},
	}, nil
}
