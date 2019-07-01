package common

import (
	"bytes"
	"context"
	"fmt"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	corev1 "k8s.io/api/core/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"strings"
	"text/template"
)

const (
	RBACConfig               = "rbac"
	StorageHostLabels        = "storage.zcloud.cn/storagetype"
	StorageBlocksAnnotations = "storage.zcloud.cn/blocks"
	StorageNamespace         = "zcloud"
	NodeIPLabels             = "zdnscloud.cn/internal-ip"
	StorageHostRole          = "node-role.kubernetes.io/storage"
)

var ctx = context.TODO()

func CreateNodeAnnotationsAndLabels(cli client.Client, cluster *storagev1.Cluster) error {
	for _, host := range cluster.Spec.Hosts {
		log.Debugf("Add Annotations and Labels for host:%s", host.NodeName)
		node := corev1.Node{}
		if err := cli.Get(ctx, k8stypes.NamespacedName{"", host.NodeName}, &node); err != nil {
			return err
		}
		//node.Labels[StorageHostLabels] = nodelabelvalue
		node.Labels[StorageHostRole] = "true"
		node.Annotations[StorageBlocksAnnotations] = strings.Replace(strings.Trim(fmt.Sprint(host.BlockDevices), "[]"), " ", ",", -1)
		if cluster.Spec.StorageType == "lvm" {
			node.Labels[StorageHostLabels] = "Lvm"
		}
		if err := cli.Update(ctx, &node); err != nil {
			return err
		}
	}
	return nil
}

func UpdateNodeAnnotations(cli client.Client, cluster *storagev1.Cluster) error {
	for _, host := range cluster.Spec.Hosts {
		log.Debugf("Update Annotations for host:%s", host.NodeName)
		node := corev1.Node{}
		if err := cli.Get(ctx, k8stypes.NamespacedName{"", host.NodeName}, &node); err != nil {
			return err
		}
		node.Annotations[StorageBlocksAnnotations] = strings.Replace(strings.Trim(fmt.Sprint(host.BlockDevices), "[]"), " ", ",", -1)
		return cli.Update(ctx, &node)
	}
	return nil
}

func CompileTemplateFromMap(tmplt string, configMap interface{}) (string, error) {
	out := new(bytes.Buffer)
	t := template.Must(template.New("compiled_template").Parse(tmplt))
	if err := t.Execute(out, configMap); err != nil {
		return "", err
	}
	return out.String(), nil
}

func DeleteNodeAnnotationsAndLabels(cli client.Client, cluster *storagev1.Cluster) error {
	for _, host := range cluster.Spec.Hosts {
		log.Debugf("Del Annotations and Labels for host:%s", host.NodeName)
		node := corev1.Node{}
		if err := cli.Get(ctx, k8stypes.NamespacedName{"", host.NodeName}, &node); err != nil {
			return err
		}
		//delete(node.Labels, StorageHostLabels)
		delete(node.Labels, StorageHostRole)
		delete(node.Annotations, StorageBlocksAnnotations)
		if cluster.Spec.StorageType == "lvm" {
			delete(node.Labels, StorageHostLabels)
		}
		if err := cli.Update(ctx, &node); err != nil {
			return err
		}
	}
	return nil
}

func GetHostAddr(cli client.Client, name string) (string, error) {
	node := corev1.Node{}
	if err := cli.Get(ctx, k8stypes.NamespacedName{"", name}, &node); err != nil {
		return "", err
	}
	return node.Annotations[NodeIPLabels], nil
}

func MakeClusterCfg(cfg map[string][]string, nodeLabelValue string) *storagev1.Cluster {
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
			StorageType: nodeLabelValue,
			Hosts:       hosts,
		},
	}
}
