package util

import (
	"context"
	"encoding/json"
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/common"
	zketypes "github.com/zdnscloud/zke/types"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8sstoragev1 "k8s.io/api/storage/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"os"
	"os/exec"
	"path"
	"strings"
)

var ctx = context.TODO()
var root = "/etc/ceph"
var files = []string{"ceph.conf", "ceph.client.admin.keyring", "ceph.mon.keyring"}

func CheckConfigMap(cli client.Client, namespace, name string) (bool, error) {
	cm := corev1.ConfigMap{}
	err := cli.Get(ctx, k8stypes.NamespacedName{namespace, name}, &cm)
	if err != nil {
		return false, err
	}
	return true, nil
}

func CheckSecret(cli client.Client, namespace, name string) (bool, error) {
	secret := corev1.Secret{}
	err := cli.Get(ctx, k8stypes.NamespacedName{namespace, name}, &secret)
	if err != nil {
		return false, err
	}
	return true, nil
}

func CheckStorageclass(cli client.Client, name string) (bool, error) {
	sc := k8sstoragev1.StorageClass{}
	err := cli.Get(ctx, k8stypes.NamespacedName{"", name}, &sc)
	if err != nil {
		return false, err
	}
	return true, nil
}

func CheckPodPhase(cli client.Client, name, stat string) (bool, error) {
	pods := corev1.PodList{}
	err := cli.List(ctx, &client.ListOptions{Namespace: common.StorageNamespace}, &pods)
	if err != nil {
		return false, err
	}
	for _, p := range pods.Items {
		if strings.Contains(p.Name, name) && string(p.Status.Phase) == stat {
			return true, nil
		}
	}
	return false, nil
}

func CheckPodDel(cli client.Client, name string) (bool, error) {
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

func ToSlice(cluster storagev1.Cluster) []string {
	infos := make([]string, 0)
	for _, host := range cluster.Status.Config {
		for _, dev := range host.BlockDevices {
			info := host.NodeName + ":" + dev
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

func GetMonDpReadyNum(cli client.Client, namespace, name string) (int32, error) {
	var num int32
	deploy := appsv1.Deployment{}
	err := cli.Get(ctx, k8stypes.NamespacedName{namespace, name}, &deploy)
	if err != nil {
		return num, err
	}
	return deploy.Status.ReadyReplicas, nil
}

func GetOsdDsReadyNum(cli client.Client, namespace, name string) (int32, error) {
	var num int32
	daemonSet := appsv1.DaemonSet{}
	err := cli.Get(context.TODO(), k8stypes.NamespacedName{namespace, name}, &daemonSet)
	if err != nil {
		return num, err
	}
	return daemonSet.Status.NumberReady, nil
}

func GetMonEp(cli client.Client) (map[string]string, error) {
	svc := make(map[string]string)
	ep := corev1.Endpoints{}
	err := cli.Get(ctx, k8stypes.NamespacedName{common.StorageNamespace, global.MonSvc}, &ep)
	if err != nil {
		return svc, err
	}
	for _, sub := range ep.Subsets {
		for _, ads := range sub.Addresses {
			svc[ads.TargetRef.Name] = ads.IP
		}
	}
	return svc, nil
}

func GetCephUUID(cli client.Client) (string, error) {
	storageclusters := storagev1.ClusterList{}
	err := cli.List(ctx, nil, &storageclusters)
	if err != nil {
		return "", err
	}
	for _, sc := range storageclusters.Items {
		if sc.Spec.StorageType != "ceph" {
			continue
		}
		return string(sc.UID), nil
	}
	return "", nil
}

func RemoveConf(cli client.Client) error {
	for _, f := range files {
		file := path.Join(root, f)
		_, err := ExecCMDWithOutput("rm", []string{"-f", file})
		if err != nil {
			return err
		}
	}
	return nil
}
func SaveConf(cli client.Client) error {
	cm := corev1.ConfigMap{}
	err := cli.Get(ctx, k8stypes.NamespacedName{common.StorageNamespace, global.ConfigMapName}, &cm)
	if err != nil {
		return err
	}
	for _, f := range files {
		date := []byte(cm.Data[f])
		file := path.Join(root, f)
		if err := ioutil.WriteFile(file, date, 0644); err != nil {
			return err
		}
	}
	return nil
}

func CheckConf() bool {
	file := path.Join(root, "ceph.conf")
	_, err := os.Stat(file)
	if err != nil {
		return false
	}
	return true
}
