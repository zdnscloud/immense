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
)

func AssembleCreateConfig(cli client.Client, cluster *storagev1.Cluster) (storagev1.Cluster, error) {
	//hosts, err := GetInfosFromStoragecluster(cli, cluster.Name)
	storagecluster, err := GetStorage(cli, cluster.Name)
	if err != nil {
		return *cluster, err
	}
	infos := make([]storagev1.HostInfo, 0)
	for _, h := range cluster.Spec.Hosts {
		devs := make([]storagev1.Dev, 0)
		exist, devstmp := isExist(h, storagecluster.Status.Config)
		if !exist {
			devstmp, err := GetBlocksFromClusterAgent(cli, h)
			if err != nil {
				return *cluster, err
			}
			devs = append(devs, devstmp...)
		} else {
			devs = append(devs, devstmp...)
		}

		info := storagev1.HostInfo{
			NodeName:     h,
			BlockDevices: devs,
		}
		infos = append(infos, info)
	}
	if err := UpdateStorageclusterConfig(cli, cluster.Name, "add", infos); err != nil {
		return *cluster, err
	}
	cluster.Status.Config = infos
	return *cluster, nil
}

func isExist(h string, infos []storagev1.HostInfo) (bool, []storagev1.Dev) {
	for _, info := range infos {
		if info.NodeName == h {
			return true, info.BlockDevices
		}
	}
	return false, []storagev1.Dev{}
}

func UpdateStorageclusterConfig(cli client.Client, name, action string, infos []storagev1.HostInfo) error {
	//ctx := context.TODO()
	storagecluster, err := GetStorage(cli, name)
	/*
		storagecluster := storagev1.Cluster{}
		err := cli.Get(ctx, k8stypes.NamespacedName{"", name}, &storagecluster)
	*/
	if err != nil {
		return err
	}
	oldinfos := storagecluster.Status.Config
	newinfos := make([]storagev1.HostInfo, 0)
	hosts := make([]string, 0)
	for _, host := range infos {
		hosts = append(hosts, host.NodeName)
	}
	if action == "add" {
		for _, h := range infos {
			var exist bool
			for _, v := range oldinfos {
				if h.NodeName == v.NodeName {
					exist = true
				}
			}
			if !exist {
				newinfos = append(newinfos, h)
			}
		}

		newinfos = append(newinfos, oldinfos...)
	}
	if action == "del" {
		for _, h := range infos {
			for i, v := range oldinfos {
				if h.NodeName != v.NodeName {
					continue
				}
				oldinfos = append(oldinfos[:i], oldinfos[i+1:]...)
			}
		}
		newinfos = append(newinfos, oldinfos...)
	}
	storagecluster.Status.Config = newinfos
	return cli.Update(ctx, &storagecluster)
}

func AssembleDeleteConfig(cli client.Client, cluster *storagev1.Cluster) (storagev1.Cluster, error) {
	hosts := make([]storagev1.HostInfo, 0)
	for _, h := range cluster.Spec.Hosts {
		/*
			var host Host
			host.NodeName = h
			devs, err := GetBlocksFromAnnotation(cli, h)
		*/
		//infos, err := GetInfosFromStoragecluster(cli, cluster.Name)
		storagecluster, err := GetStorage(cli, cluster.Name)
		if err != nil {
			return *cluster, err
		}
		for _, info := range storagecluster.Status.Config {
			if info.NodeName != h {
				continue
			}
			hosts = append(hosts, info)
		}
	}
	if err := UpdateStorageclusterConfig(cli, cluster.Name, "del", hosts); err != nil {
		return *cluster, err
	}
	cluster.Status.Config = hosts
	return *cluster, nil
}

/*
func GetInfosFromStoragecluster(cli client.Client, name string) ([]storagev1.HostInfo, error) {
	GetStorage(cli, name)
	infos := make([]storagev1.HostInfo, 0)
	storagecluster := storagev1.Cluster{}
	err := cli.Get(context.TODO(), k8stypes.NamespacedName{"", name}, &storagecluster)
	if err != nil {
		return infos, err
	}
	return storagecluster.Status.Config, nil
}*/

func AssembleUpdateConfig(cli client.Client, oldc, newc *storagev1.Cluster) (storagev1.Cluster, storagev1.Cluster, error) {
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
		return *oldc, *newc, err
	}
	adds, err := AssembleCreateConfig(cli, addc)
	if err != nil {
		return *oldc, *newc, err
	}
	return dels, adds, nil
}

/*
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
}*/

func GetBlocksFromClusterAgent(cli client.Client, name string) ([]storagev1.Dev, error) {
	devs := make([]storagev1.Dev, 0)
	service := corev1.Service{}
	err := cli.Get(context.TODO(), k8stypes.NamespacedName{StorageNamespace, "cluster-agent"}, &service)
	if err != nil {
		return devs, err
	}
	url := "/apis/agent.zcloud.cn/v1/blockdevices"
	newurl := "http://" + service.Spec.ClusterIP + url
	req, err := http.NewRequest("GET", newurl, nil)
	if err != nil {
		return devs, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return devs, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var info Data
	json.Unmarshal(body, &info)
	for _, h := range info.Data {
		if h.NodeName != name {
			continue
		}
		for _, d := range h.BlockDevices {
			if d.Parted || d.Filesystem || d.Mount {
				continue
			}
			dev := storagev1.Dev{
				Name: d.Name,
				Size: d.Size,
			}
			devs = append(devs, dev)
		}
	}
	return devs, nil
}

/*
func GetStorage(cli client.Client, name string) (storagev1.Cluster, error) {
	storagecluster := storagev1.Cluster{}
	err := cli.Get(context.TODO(), k8stypes.NamespacedName{"", name}, &storagecluster)
	if err != nil {
		return storagecluster, err
	}
	return storagecluster, nil
}*/
