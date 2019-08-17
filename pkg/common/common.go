package common

import (
	"bytes"
	"context"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	corev1 "k8s.io/api/core/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
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

func CreateNodeAnnotationsAndLabels(cli client.Client, cluster storagev1.Cluster) {
	for _, host := range cluster.Spec.Hosts {
		log.Debugf("[%s] Add Labels for host:%s", cluster.Spec.StorageType, host)
		node := corev1.Node{}
		if err := cli.Get(ctx, k8stypes.NamespacedName{"", host}, &node); err != nil {
			log.Warnf("[%s] Add Labels for host %s failed. Err: %s", cluster.Spec.StorageType, host, err.Error())
			continue
		}
		node.Labels[StorageHostRole] = "true"
		switch cluster.Spec.StorageType {
		case "lvm":
			node.Labels[StorageHostLabels] = LvmLabelsValue
		case "ceph":
			node.Labels[StorageHostLabels] = CephLabelsValue
		}
		if err := cli.Update(ctx, &node); err != nil {
			log.Warnf("[%s] Add Labels for host %s failed. Err: %s", cluster.Spec.StorageType, host, err.Error())
			continue
		}
	}
}

func DeleteNodeAnnotationsAndLabels(cli client.Client, cluster storagev1.Cluster) {
	for _, host := range cluster.Spec.Hosts {
		log.Debugf("[%s] Del Labels for host:%s", cluster.Spec.StorageType, host)
		node := corev1.Node{}
		if err := cli.Get(ctx, k8stypes.NamespacedName{"", host}, &node); err != nil {
			log.Warnf("[%s] Del Labels for host %s failed. Err: %s", cluster.Spec.StorageType, host, err.Error())
			continue
		}
		delete(node.Labels, StorageHostRole)
		delete(node.Labels, StorageHostLabels)
		if err := cli.Update(ctx, &node); err != nil {
			log.Warnf("[%s] Del Labels for host %s failed. Err: %s", cluster.Spec.StorageType, host, err.Error())
			continue
		}
	}
}

func CompileTemplateFromMap(tmplt string, configMap interface{}) (string, error) {
	out := new(bytes.Buffer)
	t := template.Must(template.New("compiled_template").Parse(tmplt))
	if err := t.Execute(out, configMap); err != nil {
		return "", err
	}
	return out.String(), nil
}

func UpdateStatus(cli client.Client, name string, phase string, message string, capacity storagev1.Capacity) error {
	storagecluster := storagev1.Cluster{}
	err := cli.Get(ctx, k8stypes.NamespacedName{"", name}, &storagecluster)
	if err != nil {
		return err
	}
	storagecluster.Status.Phase = phase
	storagecluster.Status.Message = message
	storagecluster.Status.Capacity = capacity
	return cli.Update(ctx, &storagecluster)
}

func UpdateStatusPhase(cli client.Client, name, phase string) {
	storagecluster := storagev1.Cluster{}
	if err := cli.Get(ctx, k8stypes.NamespacedName{"", name}, &storagecluster); err != nil {
		log.Warnf("Update storage cluster %s status failed. Err: %s", name, err.Error())
		return
	}
	storagecluster.Status.Phase = phase
	if err := cli.Update(ctx, &storagecluster); err != nil {
		log.Warnf("Update storage cluster %s status failed. Err: %s", name, err.Error())
	}
}

func GetStorage(cli client.Client, name string) (storagev1.Cluster, error) {
	storagecluster := storagev1.Cluster{}
	err := cli.Get(ctx, k8stypes.NamespacedName{"", name}, &storagecluster)
	if err != nil {
		return storagecluster, err
	}
	return storagecluster, nil
}
