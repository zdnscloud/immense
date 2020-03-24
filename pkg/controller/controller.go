package controller

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/cement/slice"
	"github.com/zdnscloud/gok8s/cache"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/controller"
	"github.com/zdnscloud/gok8s/event"
	"github.com/zdnscloud/gok8s/handler"
	"github.com/zdnscloud/gok8s/predicate"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	cephGlobal "github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/cluster"
	"github.com/zdnscloud/immense/pkg/common"
	"github.com/zdnscloud/immense/pkg/iscsi"
	"github.com/zdnscloud/immense/pkg/lvm"
	"github.com/zdnscloud/immense/pkg/nfs"
)

var ctx = context.TODO()

type Controller struct {
	stopCh     chan struct{}
	clusterMgr *cluster.HandlerManager
	iscsiMgr   *iscsi.HandlerManager
	nfsMgr     *nfs.HandlerManager
	client     client.Client
}

func New(config *rest.Config) (*Controller, error) {
	scm := scheme.Scheme
	storagev1.AddToScheme(scm)

	opts := cache.Options{
		Scheme: scm,
	}
	c, err := cache.New(config, opts)
	if err != nil {
		return nil, err
	}
	var options client.Options
	options.Scheme = client.GetDefaultScheme()
	storagev1.AddToScheme(options.Scheme)

	cli, err := client.New(config, options)
	if err != nil {
		return nil, err
	}

	stopCh := make(chan struct{})
	go c.Start(stopCh)
	c.WaitForCacheSync(stopCh)

	storageCtrl := &Controller{
		stopCh:     stopCh,
		clusterMgr: cluster.New(cli),
		iscsiMgr:   iscsi.New(cli),
		nfsMgr:     nfs.New(cli),
		client:     cli,
	}
	ctrl := controller.New("zcloudStorage", c, scm)
	ctrl.Watch(&storagev1.Cluster{})
	ctrl.Watch(&storagev1.Iscsi{})
	ctrl.Watch(&storagev1.Nfs{})
	ctrl.Watch(&corev1.PersistentVolumeClaim{})
	ctrl.Watch(&corev1.Secret{})
	ctrl.Start(stopCh, storageCtrl, predicate.NewIgnoreUnchangedUpdate())
	return storageCtrl, nil
}

func (d *Controller) OnCreate(e event.CreateEvent) (handler.Result, error) {
	switch obj := e.Object.(type) {
	case *storagev1.Cluster:
		go func() {
			if err := d.clusterMgr.Create(obj); err != nil {
				log.Warnf("Create failed:%s", err.Error())
			}
		}()
	case *storagev1.Iscsi:
		go func() {
			if err := d.iscsiMgr.Create(obj); err != nil {
				log.Warnf("Create failed:%s", err.Error())
			}
		}()
	case *storagev1.Nfs:
		go func() {
			if err := d.nfsMgr.Create(obj); err != nil {
				log.Warnf("Create failed:%s", err.Error())
			}
		}()
	case *corev1.PersistentVolumeClaim:
		go func() {
			if obj.Spec.StorageClassName != nil {
				if err := d.CreateFinalizer(*obj.Spec.StorageClassName); err != nil {
					log.Warnf("Watch to persistentVolumeClaim create event, but add finalizer failed:%s", err.Error())
				}
			}
		}()
	}
	return handler.Result{}, nil
}

func (d *Controller) OnUpdate(e event.UpdateEvent) (handler.Result, error) {
	switch e.ObjectOld.(type) {
	case *storagev1.Cluster:
		go func() {
			oldc := e.ObjectOld.(*storagev1.Cluster)
			newc := e.ObjectNew.(*storagev1.Cluster)
			if !reflect.DeepEqual(oldc.Spec, newc.Spec) {
				if err := d.clusterMgr.Update(oldc, newc); err != nil {
					log.Warnf("Update failed:%s", err.Error())
				}
				return
			}
			if newc.DeletionTimestamp != nil &&
				reflect.DeepEqual(oldc.Status, newc.Status) &&
				slice.SliceIndex(newc.Finalizers, common.StorageInUsedFinalizer) == -1 {
				if err := d.clusterMgr.Delete(newc); err != nil {
					log.Warnf("Delete failed:%s", err.Error())
				}
				return
			}
		}()
	case *storagev1.Iscsi:
		go func() {
			oldc := e.ObjectOld.(*storagev1.Iscsi)
			newc := e.ObjectNew.(*storagev1.Iscsi)
			if !reflect.DeepEqual(oldc.Spec, newc.Spec) {
				if err := d.iscsiMgr.Update(oldc, newc); err != nil {
					log.Warnf("Update failed:%s", err.Error())
				}
				return
			}
			if newc.DeletionTimestamp != nil &&
				reflect.DeepEqual(oldc.Status, newc.Status) &&
				slice.SliceIndex(newc.Finalizers, common.StorageInUsedFinalizer) == -1 {
				if err := d.iscsiMgr.Delete(newc); err != nil {
					log.Warnf("Delete failed:%s", err.Error())
				}
				return
			}
		}()
	case *storagev1.Nfs:
		go func() {
			oldc := e.ObjectOld.(*storagev1.Nfs)
			newc := e.ObjectNew.(*storagev1.Nfs)
			if !reflect.DeepEqual(oldc.Spec, newc.Spec) {
				if err := d.nfsMgr.Update(oldc, newc); err != nil {
					log.Warnf("Update failed:%s", err.Error())
				}
				return
			}
			if newc.DeletionTimestamp != nil &&
				reflect.DeepEqual(oldc.Status, newc.Status) &&
				slice.SliceIndex(newc.Finalizers, common.StorageInUsedFinalizer) == -1 {
				if err := d.nfsMgr.Delete(newc); err != nil {
					log.Warnf("Delete failed:%s", err.Error())
				}
				return
			}
		}()
	case *corev1.Secret:
		go func() {
			newc := e.ObjectNew.(*corev1.Secret)
			if strings.HasSuffix(newc.Name, iscsi.IscsiInstanceSecretSuffix) && newc.Namespace == common.StorageNamespace {
				iscsiName := strings.TrimRight(newc.Name, "-"+iscsi.IscsiInstanceSecretSuffix)
				if err := d.iscsiMgr.Redeploy(iscsiName); err != nil {
					log.Warnf("Update failed:%s", err.Error())
				}
			}
		}()
	}
	return handler.Result{}, nil
}

func (d *Controller) OnDelete(e event.DeleteEvent) (handler.Result, error) {
	switch obj := e.Object.(type) {
	case *corev1.PersistentVolumeClaim:
		go func() {
			if obj.Spec.StorageClassName != nil {
				if err := d.DeleteFinalizer(*obj.Spec.StorageClassName); err != nil {
					log.Warnf("Watch to persistentVolumeClaim delete event, but delete finalizer failed:%s", err.Error())
				}
			}
		}()
	}
	return handler.Result{}, nil
}

func (d *Controller) OnGeneric(e event.GenericEvent) (handler.Result, error) {
	return handler.Result{}, nil
}

func (d *Controller) CreateFinalizer(storageclass string) error {
	driver, err := common.GetProvisionerFromStorageclass(d.client, storageclass)
	if err != nil {
		return err
	}
	switch {
	case strings.HasSuffix(driver, lvm.LvmDriverSuffix):
		name := strings.TrimSuffix(driver, fmt.Sprintf(".%s", lvm.LvmDriverSuffix))
		return common.AddFinalizerForStorage(d.client, name, common.StorageInUsedFinalizer)
	case strings.HasSuffix(driver, cephGlobal.CephFsDriverSuffix):
		name := strings.TrimSuffix(driver, fmt.Sprintf(".%s", cephGlobal.CephFsDriverSuffix))
		return common.AddFinalizerForStorage(d.client, name, common.StorageInUsedFinalizer)
	case strings.HasSuffix(driver, iscsi.IscsiDriverSuffix):
		name := strings.TrimSuffix(driver, fmt.Sprintf(".%s", iscsi.IscsiDriverSuffix))
		return iscsi.AddFinalizer(d.client, name, common.StorageInUsedFinalizer)
	case strings.HasSuffix(driver, nfs.NfsDriverSuffix):
		name := strings.TrimSuffix(driver, fmt.Sprintf(".%s", nfs.NfsDriverSuffix))
		return nfs.AddFinalizer(d.client, name, common.StorageInUsedFinalizer)
	}
	return nil
}

func (d *Controller) DeleteFinalizer(storageclass string) error {
	driver, err := common.GetProvisionerFromStorageclass(d.client, storageclass)
	if err != nil {
		return err
	}
	lastone, err := common.IsPvcLastOne(d.client, driver)
	if err != nil {
		return err
	}
	if !lastone {
		return nil
	}
	switch {
	case strings.HasSuffix(driver, lvm.LvmDriverSuffix):
		name := strings.TrimSuffix(driver, fmt.Sprintf(".%s", lvm.LvmDriverSuffix))
		return common.DelFinalizerForStorage(d.client, name, common.StorageInUsedFinalizer)
	case strings.HasSuffix(driver, cephGlobal.CephFsDriverSuffix):
		name := strings.TrimSuffix(driver, fmt.Sprintf(".%s", cephGlobal.CephFsDriverSuffix))
		return common.DelFinalizerForStorage(d.client, name, common.StorageInUsedFinalizer)
	case strings.HasSuffix(driver, iscsi.IscsiDriverSuffix):
		name := strings.TrimSuffix(driver, fmt.Sprintf(".%s", iscsi.IscsiDriverSuffix))
		return iscsi.RemoveFinalizer(d.client, name, common.StorageInUsedFinalizer)
	case strings.HasSuffix(driver, nfs.NfsDriverSuffix):
		name := strings.TrimSuffix(driver, fmt.Sprintf(".%s", nfs.NfsDriverSuffix))
		return nfs.RemoveFinalizer(d.client, name, common.StorageInUsedFinalizer)
	}
	return nil
}
