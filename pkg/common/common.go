package common

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"text/template"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8sstoragev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	k8stypes "k8s.io/apimachinery/pkg/types"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
)

const (
	StorageInUsedFinalizer      = "storage.zcloud.cn/inused"
	StoragePrestopHookFinalizer = "storage.zcloud.cn/prestophook"
	PvAnnotationsKey            = "volume.beta.kubernetes.io/storage-provisioner"
	RBACConfig                  = "rbac"
	StorageHostRole             = "node-role.kubernetes.io/storage"
	StorageHostLabels           = "storage.zcloud.cn/storagetype"
	StorageNamespace            = "zcloud"
	PodCheckInterval            = 10
	LvmdPort                    = "1736"
)

var (
	ctx = context.TODO()
)

func StatefulSetObj() *appsv1.StatefulSet {
	return &appsv1.StatefulSet{}
}

func DaemonSetObj() *appsv1.DaemonSet {
	return &appsv1.DaemonSet{}
}

func DeploymentObj() *appsv1.Deployment {
	return &appsv1.Deployment{}
}

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
		return err
	}
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
		return err
	}
	return nil
}

func IsPvcLastOne(cli client.Client, driver string) (bool, error) {
	namespaces := corev1.NamespaceList{}
	if err := cli.List(context.TODO(), nil, &namespaces); err != nil {
		return false, err
	}
	for _, namespace := range namespaces.Items {
		pvcs := corev1.PersistentVolumeClaimList{}
		if err := cli.List(context.TODO(), &client.ListOptions{Namespace: namespace.Name}, &pvcs); err != nil {
			return false, err
		}
		for _, pvc := range pvcs.Items {
			if _driver, ok := pvc.Annotations[PvAnnotationsKey]; ok && _driver == driver {
				return false, nil
			}
		}
	}
	return true, nil
}

func IsObjReady(obj runtime.Object, cli client.Client, namespace, name string) (bool, error) {
	err := cli.Get(ctx, k8stypes.NamespacedName{namespace, name}, obj)
	if err != nil {
		return false, err
	}
	switch r := obj.(type) {
	case *appsv1.StatefulSet:
		if *r.Spec.Replicas == 0 {
			return false, nil
		}
		fmt.Println("sts", name, *r.Spec.Replicas, r.Status.ReadyReplicas)
		return r.Status.ReadyReplicas == *r.Spec.Replicas, nil
	case *appsv1.DaemonSet:
		if r.Status.DesiredNumberScheduled == 0 {
			return false, nil
		}
		fmt.Println("ds", name, r.Status.DesiredNumberScheduled, r.Status.NumberReady)
		return r.Status.NumberReady == r.Status.DesiredNumberScheduled, nil
	case *appsv1.Deployment:
		if *r.Spec.Replicas == 0 {
			return false, nil
		}
		fmt.Println("dp", name, *r.Spec.Replicas, r.Status.ReadyReplicas)
		return r.Status.ReadyReplicas == *r.Spec.Replicas, nil
	}
	return false, errors.New("unknow runtime object")
}

func IsObjTerminated(obj runtime.Object, cli client.Client, namespace, name string) (bool, error) {
	err := cli.Get(ctx, k8stypes.NamespacedName{namespace, name}, obj)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

func WaitTerminated(obj runtime.Object, cli client.Client, namespace, name string) error {
	log.Debugf("Wait %s terminated, this will take some time", name)
	for {
		terminated, err := IsObjTerminated(obj, cli, namespace, name)
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

func WaitReady(obj runtime.Object, cli client.Client, namespace, name string) error {
	log.Debugf("Wait %s ready, this will take some time", name)
	for {
		ready, err := IsObjReady(obj, cli, namespace, name)
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

func getDsSelector(cli client.Client, namespace, name string) (labels.Selector, error) {
	daemonSet := &appsv1.DaemonSet{}
	if err := cli.Get(ctx, k8stypes.NamespacedName{namespace, name}, daemonSet); err != nil {
		return nil, err
	}
	return metav1.LabelSelectorAsSelector(daemonSet.Spec.Selector)
}

func GetProvisionerFromStorageclass(cli client.Client, name string) (string, error) {
	storageClass := k8sstoragev1.StorageClass{}
	if err := cli.Get(context.TODO(), k8stypes.NamespacedName{"", name}, &storageClass); err != nil {
		return "", err
	}
	return storageClass.Provisioner, nil
}
