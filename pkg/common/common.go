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
	LvmLabelsValue           = "Lvm"
	CephLabelsValue          = "Ceph"
	CIDRconfigMap            = "cluster-config"
	CIDRconfigMapNamespace   = "kube-system"
)

var ctx = context.TODO()

func CreateNodeAnnotationsAndLabels(cli client.Client, cluster Storage) error {
	for _, host := range cluster.Spec.Hosts {
		if len(host.BlockDevices) == 0 {
			continue
		}
		log.Debugf("[%s] Add Annotations and Labels for host:%s, devs: %s", cluster.Spec.StorageType, host.NodeName, host.BlockDevices)
		node := corev1.Node{}
		if err := cli.Get(ctx, k8stypes.NamespacedName{"", host.NodeName}, &node); err != nil {
			log.Warnf("[%s] Add Annotations and Labels for host %s. Err: %s", cluster.Spec.StorageType, host.NodeName, err.Error())
			continue
		}
		node.Labels[StorageHostRole] = "true"
		node.Annotations[StorageBlocksAnnotations] = strings.Replace(strings.Trim(fmt.Sprint(host.BlockDevices), "[]"), " ", ",", -1)
		if cluster.Spec.StorageType == "lvm" {
			node.Labels[StorageHostLabels] = LvmLabelsValue
		}
		if cluster.Spec.StorageType == "ceph" {
			node.Labels[StorageHostLabels] = CephLabelsValue
		}
		if err := cli.Update(ctx, &node); err != nil {
			log.Warnf("[%s] Add Annotations and Labels for host %s. Err: %s", cluster.Spec.StorageType, host.NodeName, err.Error())
			continue
		}
	}
	return nil
}

func DeleteNodeAnnotationsAndLabels(cli client.Client, cluster Storage) error {
	for _, host := range cluster.Spec.Hosts {
		if len(host.BlockDevices) == 0 {
			continue
		}
		log.Debugf("[%s] Del Annotations and Labels for host:%s, devs: %s", cluster.Spec.StorageType, host.NodeName, host.BlockDevices)
		node := corev1.Node{}
		if err := cli.Get(ctx, k8stypes.NamespacedName{"", host.NodeName}, &node); err != nil {
			log.Warnf("[%s] Del Annotations and Labels for host %s. Err: %s", cluster.Spec.StorageType, host.NodeName, err.Error())
			continue
		}
		delete(node.Labels, StorageHostRole)
		delete(node.Annotations, StorageBlocksAnnotations)
		delete(node.Labels, StorageHostLabels)
		if err := cli.Update(ctx, &node); err != nil {
			log.Warnf("[%s] Del Annotations and Labels for host %s. Err: %s", cluster.Spec.StorageType, host.NodeName, err.Error())
			continue
		}
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

func getHostAddr(cli client.Client, name string) (string, error) {
	node := corev1.Node{}
	if err := cli.Get(ctx, k8stypes.NamespacedName{"", name}, &node); err != nil {
		return "", err
	}
	return node.Annotations[NodeIPLabels], nil
}

func MakeClusterCfg(cfg map[string][]string, nodeLabelValue string) Storage {
	hosts := make([]Host, 0)
	for k, v := range cfg {
		host := Host{
			NodeName:     k,
			BlockDevices: v,
		}
		hosts = append(hosts, host)
	}
	return Storage{
		Spec: StorageSpec{
			StorageType: nodeLabelValue,
			Hosts:       hosts,
		},
	}
}

func UpdateStatus(cli client.Client, name string, status storagev1.ClusterStatus) error {
	storagecluster := storagev1.Cluster{}
	err := cli.Get(ctx, k8stypes.NamespacedName{"", name}, &storagecluster)
	if err != nil {
		return err
	}
	storagecluster.Status = status
	return cli.Update(ctx, &storagecluster)
}

func UpdateStatusPhase(cli client.Client, name, phase string) error {
	storagecluster := storagev1.Cluster{}
	err := cli.Get(ctx, k8stypes.NamespacedName{"", name}, &storagecluster)
	if err != nil {
		return err
	}
	storagecluster.Status.Phase = phase
	return cli.Update(ctx, &storagecluster)
}
