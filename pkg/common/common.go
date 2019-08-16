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

func CreateNodeAnnotationsAndLabels(cli client.Client, cluster storagev1.Cluster) error {
	for _, host := range cluster.Spec.Hosts {
		log.Debugf("[%s] Add Labels for host:%s", cluster.Spec.StorageType, host)
		node := corev1.Node{}
		if err := cli.Get(ctx, k8stypes.NamespacedName{"", host}, &node); err != nil {
			log.Warnf("[%s] Add Labels for host %s. Err: %s", cluster.Spec.StorageType, host, err.Error())
			continue
		}
		node.Labels[StorageHostRole] = "true"
		//	node.Annotations[StorageBlocksAnnotations] = strings.Replace(strings.Trim(fmt.Sprint(host.BlockDevices), "[]"), " ", ",", -1)
		if cluster.Spec.StorageType == "lvm" {
			node.Labels[StorageHostLabels] = LvmLabelsValue
		}
		if cluster.Spec.StorageType == "ceph" {
			node.Labels[StorageHostLabels] = CephLabelsValue
		}
		if err := cli.Update(ctx, &node); err != nil {
			log.Warnf("[%s] Add Annotations and Labels for host %s. Err: %s", cluster.Spec.StorageType, host, err.Error())
			continue
		}
	}
	return nil
}

func DeleteNodeAnnotationsAndLabels(cli client.Client, cluster storagev1.Cluster) error {
	for _, host := range cluster.Spec.Hosts {
		log.Debugf("[%s] Del Labels for host:%s", cluster.Spec.StorageType, host)
		node := corev1.Node{}
		if err := cli.Get(ctx, k8stypes.NamespacedName{"", host}, &node); err != nil {
			log.Warnf("[%s] Del Labels for host %s. Err: %s", cluster.Spec.StorageType, host, err.Error())
			continue
		}
		delete(node.Labels, StorageHostRole)
		//delete(node.Annotations, StorageBlocksAnnotations)
		delete(node.Labels, StorageHostLabels)
		if err := cli.Update(ctx, &node); err != nil {
			log.Warnf("[%s] Del Labels for host %s. Err: %s", cluster.Spec.StorageType, host, err.Error())
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

/*
func getHostAddr(cli client.Client, name string) (string, error) {
	node := corev1.Node{}
	if err := cli.Get(ctx, k8stypes.NamespacedName{"", name}, &node); err != nil {
		return "", err
	}
	return node.Annotations[NodeIPLabels], nil
}*/

/*
func MakeClusterCfg(cfg map[string][]string, nodeLabelValue string) storagev1.Cluster {
	hosts := make([]storagev1.HostInfo, 0)
	for k, v := range cfg {
		devs := make([]storagev1.Dev, 0)
		for _, d := range v {
			dev := storagev1.Dev{
				Name: d,
			}
			devs = append(devs, dev)
		}
		host := storagev1.HostInfo{
			NodeName:     k,
			BlockDevices: devs,
		}
		hosts = append(hosts, host)
	}
	return storagev1.Cluster{
		Spec: storagev1.ClusterSpec{
			StorageType: nodeLabelValue,
		},
		Status: storagev1.ClusterStatus{
			Config: hosts,
		},
	}
}*/

func UpdateStatus(cli client.Client, name string, phase string, message string, capacity storagev1.Capacity) error {
	storagecluster := storagev1.Cluster{}
	err := cli.Get(ctx, k8stypes.NamespacedName{"", name}, &storagecluster)
	if err != nil {
		return err
	}
	if phase != "" {
		storagecluster.Status.Phase = phase
	}
	if message != "" {
		storagecluster.Status.Message = message
	}
	if len(capacity.Instances) > 0 {
		storagecluster.Status.Capacity = capacity
	}
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

func GetStorage(cli client.Client, name string) (storagev1.Cluster, error) {
	storagecluster := storagev1.Cluster{}
	err := cli.Get(ctx, k8stypes.NamespacedName{"", name}, &storagecluster)
	if err != nil {
		return storagecluster, err
	}
	return storagecluster, nil
}
