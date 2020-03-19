package common

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8sstorage "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	k8stypes "k8s.io/apimachinery/pkg/types"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
)

const (
	StorageInUsedFinalizer      = "storage.zcloud.cn/inused"
	PvProvisionerKey            = "pv.kubernetes.io/provisioned-by"
	StoragePrestopHookFinalizer = "storage.zcloud.cn/prestophook"
	RBACConfig                  = "rbac"
	StorageHostRole             = "node-role.kubernetes.io/storage"
	StorageHostLabels           = "storage.zcloud.cn/storagetype"
	StorageNamespace            = "zcloud"
	PodCheckInterval            = 10
	LvmdPort                    = "1736"
)

var ctx = context.TODO()

func CreateNodeAnnotationsAndLabels(cli client.Client, key, value string, hosts []string) error {
	for _, host := range hosts {
		log.Debugf("Add Labels for storage %s on host:%s", value, host)
		node := corev1.Node{}
		if err := cli.Get(ctx, k8stypes.NamespacedName{"", host}, &node); err != nil {
			return fmt.Errorf("Add Labels for storage %s on host %s failed. Err: %v", value, host, err)
		}
		node.Labels[key] = value
		if err := cli.Update(ctx, &node); err != nil {
			return fmt.Errorf("Add Labels for storage %s on host %s failed. Err: %v", value, host, err)
		}
	}
	return nil
}

func DeleteNodeAnnotationsAndLabels(cli client.Client, key, value string, hosts []string) error {
	for _, host := range hosts {
		log.Debugf("Del Labels for storage %s on host:%s", value, host)
		node := corev1.Node{}
		if err := cli.Get(ctx, k8stypes.NamespacedName{"", host}, &node); err != nil {
			return fmt.Errorf("Delete Labels for storage %s on host %s failed. Err: %v", value, host, err)
		}
		delete(node.Labels, key)
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

func UpdateClusterStatus(cli client.Client, name string, status storagev1.ClusterStatus) error {
	storagecluster, err := GetStorageCluster(cli, name)
	if err != nil {
		return err
	}
	storagecluster.Status = status
	return cli.Update(ctx, storagecluster)
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

func GetStorageClusterFromPv(cli client.Client, name string) (string, error) {
	pv := &corev1.PersistentVolume{}
	if err := cli.Get(context.TODO(), k8stypes.NamespacedName{"", name}, pv); err != nil {
		return "", err
	}
	return pv.Spec.StorageClassName, nil
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

func IsVaLastOne(cli client.Client, driver string) (bool, error) {
	volumeattachments := k8sstorage.VolumeAttachmentList{}
	if err := cli.List(ctx, nil, &volumeattachments); err != nil {
		return false, err
	}
	for _, v := range volumeattachments.Items {
		if v.Spec.Attacher == driver {
			return false, nil
		}
	}
	return true, nil
}

func IsDpReady(cli client.Client, namespace, name string) (bool, error) {
	deploy := appsv1.Deployment{}
	if err := cli.Get(ctx, k8stypes.NamespacedName{namespace, name}, &deploy); err != nil {
		return false, err
	}
	log.Debugf("Deployment: %s ready:%d, desired: %d", name, deploy.Status.ReadyReplicas, *deploy.Spec.Replicas)
	if *deploy.Spec.Replicas == 0 {
		return false, nil
	}
	return deploy.Status.ReadyReplicas == *deploy.Spec.Replicas, nil
}

func IsDsReady(cli client.Client, namespace, name string) (bool, error) {
	daemonSet := appsv1.DaemonSet{}
	if err := cli.Get(ctx, k8stypes.NamespacedName{namespace, name}, &daemonSet); err != nil {
		return false, err
	}
	log.Debugf("DaemonSet: %s ready:%d, desired: %d", name, daemonSet.Status.NumberReady, daemonSet.Status.DesiredNumberScheduled)
	if daemonSet.Status.DesiredNumberScheduled == 0 {
		return false, nil
	}
	return daemonSet.Status.NumberReady == daemonSet.Status.DesiredNumberScheduled, nil
}

func IsStsReady(cli client.Client, namespace, name string) (bool, error) {
	statefulset := appsv1.StatefulSet{}
	if err := cli.Get(ctx, k8stypes.NamespacedName{namespace, name}, &statefulset); err != nil {
		return false, err
	}
	log.Debugf("StatefulSet: %s ready:%d, desired: %d", name, statefulset.Status.ReadyReplicas, *statefulset.Spec.Replicas)
	if *statefulset.Spec.Replicas == 0 {
		return false, nil
	}
	return statefulset.Status.ReadyReplicas == *statefulset.Spec.Replicas, nil
}

func IsDpTerminated(cli client.Client, namespace, name string) (bool, error) {
	deploys := appsv1.DeploymentList{}
	if err := cli.List(ctx, &client.ListOptions{Namespace: namespace}, &deploys); err != nil {
		return false, err
	}
	for _, deploy := range deploys.Items {
		if deploy.Name == name {
			return false, nil
		}
	}
	return true, nil
}

func IsDsTerminated(cli client.Client, namespace, name string) (bool, error) {
	daemonSets := appsv1.DaemonSetList{}
	if err := cli.List(ctx, &client.ListOptions{Namespace: namespace}, &daemonSets); err != nil {
		return false, err
	}
	for _, daemonSet := range daemonSets.Items {
		if daemonSet.Name == name {
			return false, nil
		}
	}
	return true, nil
}

func IsStsTerminated(cli client.Client, namespace, name string) (bool, error) {
	statefulsets := appsv1.StatefulSetList{}
	if err := cli.List(ctx, &client.ListOptions{Namespace: namespace}, &statefulsets); err != nil {
		return false, err
	}
	for _, statefulset := range statefulsets.Items {
		if statefulset.Name == name {
			return false, nil
		}
	}
	return true, nil
}

func WaitStsTerminated(cli client.Client, namespace, name string) error {
	log.Debugf("Wait statefulset %s terminated, this will take some time", name)
	for {
		terminated, err := IsStsTerminated(cli, namespace, name)
		if err != nil {
			return err
		}
		if terminated {
			break
		}
		time.Sleep(PodCheckInterval * time.Second)
	}
	return nil
}

func WaitStsReady(cli client.Client, namespace, name string) error {
	log.Debugf("Wait statefulset %s ready, this will take some time", name)
	for {
		ready, err := IsStsReady(cli, namespace, name)
		if err != nil {
			return err
		}
		if ready {
			break
		}
		time.Sleep(PodCheckInterval * time.Second)
	}
	return nil
}

func WaitDsTerminated(cli client.Client, namespace, name string) error {
	log.Debugf("Wait daemonset %s terminated, this will take some time", name)
	for {
		terminated, err := IsDsTerminated(cli, namespace, name)
		if err != nil {
			return err
		}
		if terminated {
			break
		}
		time.Sleep(PodCheckInterval * time.Second)
	}
	return nil
}

func WaitDsReady(cli client.Client, namespace, name string) error {
	log.Debugf("Wait daemonset %s ready, this will take some time", name)
	for {
		ready, err := IsDsReady(cli, namespace, name)
		if err != nil {
			return err
		}
		if ready {
			break
		}
		time.Sleep(PodCheckInterval * time.Second)
	}
	return nil
}

func WaitDpTerminated(cli client.Client, namespace, name string) error {
	log.Debugf("Wait deployment %s terminated, this will take some time", name)
	for {
		terminated, err := IsDpTerminated(cli, namespace, name)
		if err != nil {
			return err
		}
		if terminated {
			break
		}
		time.Sleep(PodCheckInterval * time.Second)
	}
	return nil
}

func WaitDpReady(cli client.Client, namespace, name string) error {
	log.Debugf("Wait deployment %s ready, this will take some time", name)
	for {
		ready, err := IsDpReady(cli, namespace, name)
		if err != nil {
			return err
		}
		if ready {
			break
		}
		time.Sleep(PodCheckInterval * time.Second)
	}
	return nil
}

func isPodSucceeded(cli client.Client, namespace, name string) (bool, error) {
	pods := corev1.PodList{}
	err := cli.List(ctx, &client.ListOptions{Namespace: namespace}, &pods)
	if err != nil {
		return false, err
	}
	for _, pod := range pods.Items {
		if strings.Contains(pod.Name, name) && string(pod.Status.Phase) == "Succeeded" {
			return true, nil
		}
	}
	return false, nil
}

func WaitPodSucceeded(cli client.Client, namespace, name string) error {
	log.Debugf("Wait pod %s status succeeded, this will take some time", name)
	for {
		succeeded, err := isPodSucceeded(cli, namespace, name)
		if err != nil {
			return err
		}
		if succeeded {
			break
		}
		time.Sleep(PodCheckInterval * time.Second)
	}
	return nil
}

func getPods(cli client.Client, namespace string, selector labels.Selector) (*corev1.PodList, error) {
	pods := &corev1.PodList{}
	if err := cli.List(ctx, &client.ListOptions{Namespace: namespace, LabelSelector: selector}, pods); err != nil {
		return nil, err
	}
	return pods, nil
}

func getSelector(cli client.Client, namespace, name string) (labels.Selector, error) {
	daemonSet := &appsv1.DaemonSet{}
	if err := cli.Get(ctx, k8stypes.NamespacedName{namespace, name}, daemonSet); err != nil {
		return nil, err
	}
	return metav1.LabelSelectorAsSelector(daemonSet.Spec.Selector)
}
