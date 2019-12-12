package common

import (
	"bytes"
	"context"
	"strings"
	"text/template"
	"time"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8sstorage "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stypes "k8s.io/apimachinery/pkg/types"
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

func UpdateStatusPhase(cli client.Client, name string, phase storagev1.StatusPhase) {
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
	if err := cli.Get(ctx, k8stypes.NamespacedName{"", name}, &storagecluster); err != nil {
		return storagecluster, err
	}
	return storagecluster, nil
}

func AddFinalizerForStorage(cli client.Client, name, finalizer string) error {
	storagecluster, err := GetStorage(cli, name)
	if err != nil {
		return err
	}
	var obj runtime.Object
	obj = &storagecluster
	metaObj := obj.(metav1.Object)
	if helper.HasFinalizer(metaObj, finalizer) {
		return nil
	}
	helper.AddFinalizer(metaObj, finalizer)
	if err := cli.Update(ctx, obj); err != nil {
		log.Warnf("add finalizer %s for storage cluster %s failed. Err: %s", finalizer, name, err.Error())
	}
	return nil
}

func DelFinalizerForStorage(cli client.Client, name, finalizer string) error {
	storagecluster, err := GetStorage(cli, name)
	if err != nil {
		return err
	}
	var obj runtime.Object
	obj = &storagecluster
	metaObj := obj.(metav1.Object)
	if !helper.HasFinalizer(metaObj, finalizer) {
		return nil
	}
	helper.RemoveFinalizer(metaObj, finalizer)
	if err := cli.Update(ctx, obj); err != nil {
		log.Warnf("del finalizer %s for storage cluster %s failed. Err: %s", finalizer, name, err.Error())
	}
	return nil
}

/*
func GetClusterFromVolumeAttachment(cli client.Client, storageType string) (runtime.Object, error) {
	storageclusters := storagev1.ClusterList{}
	if err := cli.List(ctx, nil, &storageclusters); err != nil {
		return nil, err
	}
	for _, storage := range storageclusters.Items {
		if storage.Spec.StorageType != storageType {
			continue
		}
		var obj runtime.Object
		obj = &storage
		return obj, nil
	}
	return nil, errors.New("can not found storagecluster for type" + storageType)
}*/

func IsLastOne(cli client.Client, va *k8sstorage.VolumeAttachment) (bool, error) {
	volumeattachments := k8sstorage.VolumeAttachmentList{}
	if err := cli.List(ctx, nil, &volumeattachments); err != nil {
		return false, err
	}
	for _, v := range volumeattachments.Items {
		if v.Spec.Attacher == va.Spec.Attacher {
			return false, nil
		}
	}
	return true, nil
}

func IsDpReady(cli client.Client, namespace, name string) bool {
	deploy := appsv1.Deployment{}
	if err := cli.Get(ctx, k8stypes.NamespacedName{namespace, name}, &deploy); err != nil {
		return false
	}
	log.Debugf("Deployment: %s ready:%d, desired: %d", name, deploy.Status.ReadyReplicas, *deploy.Spec.Replicas)
	return deploy.Status.ReadyReplicas == *deploy.Spec.Replicas
}

func IsDsReady(cli client.Client, namespace, name string) bool {
	daemonSet := appsv1.DaemonSet{}
	if err := cli.Get(context.TODO(), k8stypes.NamespacedName{namespace, name}, &daemonSet); err != nil {
		return false
	}
	log.Debugf("DaemonSet: %s ready:%d, desired: %d", name, daemonSet.Status.NumberReady, daemonSet.Status.DesiredNumberScheduled)
	return daemonSet.Status.NumberReady == daemonSet.Status.DesiredNumberScheduled
}

func IsStsReady(cli client.Client, namespace, name string) bool {
	statefulset := appsv1.StatefulSet{}
	if err := cli.Get(context.TODO(), k8stypes.NamespacedName{namespace, name}, &statefulset); err != nil {
		return false
	}
	log.Debugf("StatefulSet: %s ready:%d, desired: %d", name, statefulset.Status.ReadyReplicas, *statefulset.Spec.Replicas)
	return statefulset.Status.ReadyReplicas == *statefulset.Spec.Replicas
}

func WaitCSIReady(cli client.Client, namespace, provisioner, plugin string) {
	log.Debugf("Wait all csi pod running, this will take some time")
	for {
		if !IsStsReady(cli, StorageNamespace, provisioner) || !IsDsReady(cli, StorageNamespace, plugin) {
			time.Sleep(10 * time.Second)
			continue
		}
		break
	}
}

func WaitPodTerminated(cli client.Client, namespace, prefix string) {
	log.Debugf("Wait all pods %s-xxx terminated, this will take some time")
	for {
		pods := corev1.PodList{}
		if err := cli.List(context.TODO(), &client.ListOptions{Namespace: namespace}, &pods); err != nil {
			continue
		}
		var exist bool
		for _, pod := range pods.Items {
			if strings.HasPrefix(pod.Name, prefix) {
				exist = true
				break
			}
		}
		if exist {
			time.Sleep(10 * time.Second)
			continue
		} else {
			break
		}
	}
}
