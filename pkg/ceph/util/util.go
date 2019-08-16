package util

import (
	"context"
	"encoding/json"
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/common"
	zketypes "github.com/zdnscloud/zke/types"
	corev1 "k8s.io/api/core/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"os/exec"
	"strings"
)

var ctx = context.TODO()

func GetMonIPs(cli client.Client) ([]string, error) {
	ips := make([]string, 0)
	pods := corev1.PodList{}
	err := cli.List(ctx, &client.ListOptions{Namespace: common.StorageNamespace}, &pods)
	if err != nil {
		return ips, err
	}
	for _, p := range pods.Items {
		if strings.Contains(p.Name, "ceph-mon-") && p.Status.Phase == "Running" {
			ip := p.Status.PodIP + ":6789"
			ips = append(ips, ip)
		}
	}
	return ips, nil
}

func GetMonSvc(cli client.Client) ([]string, error) {
	ips := make([]string, 0)
	ep := corev1.Endpoints{}
	err := cli.Get(ctx, k8stypes.NamespacedName{common.StorageNamespace, global.MonSvc}, &ep)
	if err != nil {
		return ips, err
	}
	for _, sub := range ep.Subsets {
		for _, ads := range sub.Addresses {
			ips = append(ips, ads.IP)
		}
	}
	return ips, nil
}

func CheckConfigMap(cli client.Client, namespace, name string) (bool, error) {
	cm := corev1.ConfigMap{}
	err := cli.Get(context.TODO(), k8stypes.NamespacedName{namespace, name}, &cm)
	if err != nil {
		return false, err
	}
	return true, nil
}

func CheckSecret(cli client.Client, namespace, name string) (bool, error) {
	secret := corev1.Secret{}
	err := cli.Get(context.TODO(), k8stypes.NamespacedName{namespace, name}, &secret)
	if err != nil {
		return false, err
	}
	return true, nil
}

/*
func GetCIDRs(cli client.Client, cluster *storagev1.Cluster) (string, error) {
	var cidrs string
	for _, host := range cluster.Spec.Hosts {
		cidr, err := getHostpodCIDR(cli, host.NodeName)
		if err != nil {
			return cidrs, err
		}
		cidrs = cidrs + cidr + ","
	}
	return strings.TrimRight(cidrs, ","), nil
}*/

func GetPodIp(cli client.Client, name string) (string, error) {
	pods := corev1.PodList{}
	err := cli.List(ctx, &client.ListOptions{Namespace: common.StorageNamespace}, &pods)
	if err != nil {
		return "", err
	}
	for _, p := range pods.Items {
		if strings.Contains(p.Name, name) && p.Status.Phase == "Running" {
			return p.Status.PodIP, nil
		}
	}
	return "", nil
}

func IsPodSucceeded(cli client.Client, name string) (bool, error) {
	pods := corev1.PodList{}
	err := cli.List(ctx, &client.ListOptions{Namespace: common.StorageNamespace}, &pods)
	if err != nil {
		return false, err
	}
	for _, p := range pods.Items {
		if strings.Contains(p.Name, name) && p.Status.Phase == "Succeeded" {
			return true, nil
		}
	}
	return false, nil
}
func IsPodDel(cli client.Client, name string) (bool, error) {
	pods := corev1.PodList{}
	err := cli.List(ctx, &client.ListOptions{Namespace: common.StorageNamespace}, &pods)
	if err != nil {
		return false, err
	}
	for _, p := range pods.Items {
		if strings.Contains(p.Name, name) {
			return false, nil
		}
	}
	return true, nil
}

func ExecCMDWithOutput(cmd string, args []string) (string, error) {
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func getHostpodCIDR(cli client.Client, name string) (string, error) {
	node := corev1.Node{}
	if err := cli.Get(ctx, k8stypes.NamespacedName{"", name}, &node); err != nil {
		return "", err
	}
	return node.Spec.PodCIDR, nil
}

func ToSlice(cluster storagev1.Cluster) []string {
	infos := make([]string, 0)
	for _, host := range cluster.Status.Config {
		for _, dev := range host.BlockDevices {
			info := host.NodeName + ":" + dev.Name
			infos = append(infos, info)
		}
	}
	return infos
}

func GetClusterCIDR(cli client.Client, namespace, name string) (string, error) {
	cm := corev1.ConfigMap{}
	err := cli.Get(ctx, k8stypes.NamespacedName{namespace, name}, &cm)
	if err != nil {
		return "", err
	}
	var res zketypes.ZKEConfig
	json.Unmarshal([]byte(cm.Data["cluster-config"]), &res)
	return res.Option.ClusterCidr, nil

}
