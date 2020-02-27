package common

import (
	"bytes"
	"context"
	"fmt"
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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

const (
	StorageInUsedFinalizer      = "storage.zcloud.cn/inused"
	StoragePrestopHookFinalizer = "storage.zcloud.cn/prestophook"
	RBACConfig                  = "rbac"
	StorageHostRole             = "node-role.kubernetes.io/storage"
	StorageHostLabels           = "storage.zcloud.cn/storagetype"
	StorageNamespace            = "zcloud"
	PodCheckInterval            = 10
)

var ctx = context.TODO()

func CreateNodeAnnotationsAndLabels(cli client.Client, value string, hosts []string) error {
	for _, host := range hosts {
		log.Debugf("Add Labels for storage %s on host:%s", value, host)
		node := corev1.Node{}
		if err := cli.Get(ctx, k8stypes.NamespacedName{"", host}, &node); err != nil {
			return fmt.Errorf("Add Labels for storage %s on host %s failed. Err: %v", value, host, err)
		}
		node.Labels[StorageHostRole] = "true"
		node.Labels[StorageHostLabels] = value
		if err := cli.Update(ctx, &node); err != nil {
			return fmt.Errorf("Add Labels for storage %s on host %s failed. Err: %v", value, host, err)
		}
	}
	return nil
}

func DeleteNodeAnnotationsAndLabels(cli client.Client, value string, hosts []string) error {
	for _, host := range hosts {
		log.Debugf("Del Labels for storage %s on host:%s", value, host)
		node := corev1.Node{}
		if err := cli.Get(ctx, k8stypes.NamespacedName{"", host}, &node); err != nil {
			return fmt.Errorf("Delete Labels for storage %s on host %s failed. Err: %v", value, host, err)
		}
		delete(node.Labels, StorageHostRole)
		delete(node.Labels, StorageHostLabels)
		if err := cli.Update(ctx, &node); err != nil {
			return fmt.Errorf("Delete Labels for storage %s on host %s failed. Err: %v", value, host, err)
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

func UpdateClusterStatusPhase(cli client.Client, name string, phase storagev1.StatusPhase) {
	storagecluster, err := GetStorageCluster(cli, name)
	if err != nil {
		if apierrors.IsNotFound(err) == true {
			return
		}
		log.Warnf("Update storage cluster %s status failed. Err: %s", name, err.Error())
		return
	}
	storagecluster.Status.Phase = phase
	if err := cli.Update(ctx, storagecluster); err != nil {
		if apierrors.IsNotFound(err) == true {
			return
		}
		log.Warnf("Update storage cluster %s status failed. Err: %s", name, err.Error())
		return
	}
	return
}

func GetStorageCluster(cli client.Client, name string) (*storagev1.Cluster, error) {
	storagecluster := &storagev1.Cluster{}
	if err := cli.Get(ctx, k8stypes.NamespacedName{"", name}, storagecluster); err != nil {
		return nil, err
	}
	return storagecluster, nil
}

func AddFinalizerForStorage(cli client.Client, name, finalizer string) error {
	storagecluster, err := GetStorageCluster(cli, name)
	if err != nil {
		return err
	}
	helper.AddFinalizer(storagecluster, finalizer)
	if err := cli.Update(ctx, storagecluster); err != nil {
		log.Warnf("add finalizer %s for storage cluster %s failed. Err: %s", finalizer, name, err.Error())
	}
	log.Debugf("Add finalizer %s for storage cluster %s", finalizer, name)
	return nil
}

func DelFinalizerForStorage(cli client.Client, name, finalizer string) error {
	storagecluster, err := GetStorageCluster(cli, name)
	if err != nil {
		if apierrors.IsNotFound(err) == true {
			return nil
		}
		return err
	}
	helper.RemoveFinalizer(storagecluster, finalizer)
	if err := cli.Update(ctx, storagecluster); err != nil {
		log.Warnf("del finalizer %s for storage cluster %s failed. Err: %s", finalizer, name, err.Error())
	}
	log.Debugf("Delete finalizer %s for storage cluster %s", finalizer, name)
	return nil
}

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
	if *deploy.Spec.Replicas == 0 {
		return false
	}
	return deploy.Status.ReadyReplicas == *deploy.Spec.Replicas
}

func IsDsReady(cli client.Client, namespace, name string) bool {
	daemonSet := appsv1.DaemonSet{}
	if err := cli.Get(ctx, k8stypes.NamespacedName{namespace, name}, &daemonSet); err != nil {
		return false
	}
	log.Debugf("DaemonSet: %s ready:%d, desired: %d", name, daemonSet.Status.NumberReady, daemonSet.Status.DesiredNumberScheduled)
	if daemonSet.Status.DesiredNumberScheduled == 0 {
		return false
	}
	return daemonSet.Status.NumberReady == daemonSet.Status.DesiredNumberScheduled
}

func IsStsReady(cli client.Client, namespace, name string) bool {
	statefulset := appsv1.StatefulSet{}
	if err := cli.Get(ctx, k8stypes.NamespacedName{namespace, name}, &statefulset); err != nil {
		return false
	}
	log.Debugf("StatefulSet: %s ready:%d, desired: %d", name, statefulset.Status.ReadyReplicas, *statefulset.Spec.Replicas)
	if *statefulset.Spec.Replicas == 0 {
		return false
	}
	return statefulset.Status.ReadyReplicas == *statefulset.Spec.Replicas
}

func IsDpTerminated(cli client.Client, namespace, name string) bool {
	deploys := appsv1.DeploymentList{}
	if err := cli.List(ctx, &client.ListOptions{Namespace: namespace}, &deploys); err != nil {
		return false
	}
	for _, deploy := range deploys.Items {
		if deploy.Name == name {
			return false
		}
	}
	return true
}

func IsDsTerminated(cli client.Client, namespace, name string) bool {
	daemonSets := appsv1.DaemonSetList{}
	if err := cli.List(ctx, &client.ListOptions{Namespace: namespace}, &daemonSets); err != nil {
		return false
	}
	for _, daemonSet := range daemonSets.Items {
		if daemonSet.Name == name {
			return false
		}
	}
	return true
}

func IsStsTerminated(cli client.Client, namespace, name string) bool {
	statefulsets := appsv1.StatefulSetList{}
	if err := cli.List(ctx, &client.ListOptions{Namespace: namespace}, &statefulsets); err != nil {
		return false
	}
	for _, statefulset := range statefulsets.Items {
		if statefulset.Name == name {
			return false
		}
	}
	return true
}

func WaitStsTerminated(cli client.Client, namespace, name string) {
	log.Debugf("Wait statefulset %s terminated, this will take some time", name)
	for {
		if !IsStsTerminated(cli, namespace, name) {
			time.Sleep(PodCheckInterval * time.Second)
			continue
		}
		break
	}
}

func WaitStsReady(cli client.Client, namespace, name string) {
	log.Debugf("Wait statefulset %s ready, this will take some time", name)
	for {
		if !IsStsReady(cli, namespace, name) {
			time.Sleep(PodCheckInterval * time.Second)
			continue
		}
		break
	}
}

func WaitDsTerminated(cli client.Client, namespace, name string) {
	log.Debugf("Wait daemonset %s terminated, this will take some time", name)
	for {
		if !IsDsTerminated(cli, namespace, name) {
			time.Sleep(PodCheckInterval * time.Second)
			continue
		}
		break
	}
}

func WaitDsReady(cli client.Client, namespace, name string) {
	log.Debugf("Wait daemonset %s ready, this will take some time", name)
	for {
		if !IsDsReady(cli, namespace, name) {
			time.Sleep(PodCheckInterval * time.Second)
			continue
		}
		break
	}
}

func WaitDpTerminated(cli client.Client, namespace, name string) {
	log.Debugf("Wait deployment %s terminated, this will take some time", name)
	for {
		if !IsDpTerminated(cli, namespace, name) {
			time.Sleep(PodCheckInterval * time.Second)
			continue
		}
		break
	}
}

func WaitDpReady(cli client.Client, namespace, name string) {
	log.Debugf("Wait deployment %s ready, this will take some time", name)
	for {
		if !IsDpReady(cli, namespace, name) {
			time.Sleep(PodCheckInterval * time.Second)
			continue
		}
		break
	}
}

func isPodSucceeded(cli client.Client, namespace, name string) bool {
	pods := corev1.PodList{}
	err := cli.List(ctx, &client.ListOptions{Namespace: namespace}, &pods)
	if err != nil {
		return false
	}
	for _, pod := range pods.Items {
		if strings.Contains(pod.Name, name) && string(pod.Status.Phase) == "Succeeded" {
			return true
		}
	}
	return false
}

func WaitPodSucceeded(cli client.Client, namespace, name string) {
	log.Debugf("Wait pod %s status succeeded, this will take some time", name)
	for {
		if !isPodSucceeded(cli, namespace, name) {
			time.Sleep(PodCheckInterval * time.Second)
			continue
		}
		break
	}
}
