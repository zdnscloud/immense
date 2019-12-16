package util

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/common"
	corev1 "k8s.io/api/core/v1"
	k8sstoragev1 "k8s.io/api/storage/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

var ctx = context.TODO()
var root = "/etc/ceph"
var files = []string{"ceph.conf", "ceph.client.admin.keyring", "ceph.mon.keyring"}

func CheckConfigMap(cli client.Client, namespace, name string) (bool, error) {
	cm := corev1.ConfigMap{}
	if err := cli.Get(ctx, k8stypes.NamespacedName{namespace, name}, &cm); err != nil {
		return false, err
	}
	return true, nil
}

func CheckSecret(cli client.Client, namespace, name string) (bool, error) {
	secret := corev1.Secret{}
	if err := cli.Get(ctx, k8stypes.NamespacedName{namespace, name}, &secret); err != nil {
		return false, err
	}
	return true, nil
}

func CheckStorageclass(cli client.Client, name string) (bool, error) {
	sc := k8sstoragev1.StorageClass{}
	if err := cli.Get(ctx, k8stypes.NamespacedName{"", name}, &sc); err != nil {
		return false, err
	}
	return true, nil
}

func ExecCMDWithOutput(cmd string, args []string) (string, error) {
	out, err := exec.Command(cmd, args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
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

func GetDevsForHost(cluster storagev1.Cluster, host string) []string {
	for _, node := range cluster.Status.Config {
		if node.NodeName != host {
			continue
		}
		return node.BlockDevices
	}
	return []string{}
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
	if err := cli.Get(ctx, k8stypes.NamespacedName{common.StorageNamespace, global.ConfigMapName}, &cm); err != nil {
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
	if _, err := os.Stat(file); err != nil {
		return false
	}
	return true
}

func getMonSvc(cli client.Client, name string) (string, error) {
	service := corev1.Service{}
	if err := cli.Get(ctx, k8stypes.NamespacedName{common.StorageNamespace, name}, &service); err != nil {
		return "", err
	}
	return service.Spec.ClusterIP, nil
}

func GetMonSvcMap(cli client.Client) (map[string]string, error) {
	svc := make(map[string]string)
	for _, id := range global.MonMembers {
		svcName := global.MonSvc + "-" + id
		addr, err := getMonSvc(cli, svcName)
		if err != nil {
			return svc, err
		}
		svc[id] = addr
	}
	return svc, nil
}

func GetCephUUID(cli client.Client) (string, error) {
	storageclusters := storagev1.ClusterList{}
	if err := cli.List(ctx, nil, &storageclusters); err != nil {
		return "", err
	}
	for _, sc := range storageclusters.Items {
		if sc.Spec.StorageType != global.StorageType {
			continue
		}
		return string(sc.UID), nil
	}
	return "", nil
}

func GetMonHosts(monsvc map[string]string) string {
	var hosts [][]string
	for _, ip := range monsvc {
		var host []string
		host1 := "v1:" + ip + ":" + global.MonPortV1
		host2 := "v2:" + ip + ":" + global.MonPortV2
		host = append(host, host2)
		host = append(host, host1)
		hosts = append(hosts, host)
	}
	return strings.Replace(strings.TrimPrefix(strings.TrimSuffix(fmt.Sprint(hosts), "]"), "["), " ", ",", -1)
}
