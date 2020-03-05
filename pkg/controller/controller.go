package controller

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

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
	k8sstorage "k8s.io/api/storage/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

var ctx = context.TODO()

type Controller struct {
	stopCh     chan struct{}
	clusterMgr *cluster.HandlerManager
	iscsiMgr   *iscsi.HandlerManager
	nfsMgr     *nfs.HandlerManager
	lock       sync.RWMutex
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
	ctrl.Watch(&k8sstorage.VolumeAttachment{})
	ctrl.Start(stopCh, storageCtrl, predicate.NewIgnoreUnchangedUpdate())
	return storageCtrl, nil
}

func (d *Controller) OnCreate(e event.CreateEvent) (handler.Result, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	switch obj := e.Object.(type) {
	case *storagev1.Cluster:
		conf := e.Object.(*storagev1.Cluster)
		if err := d.clusterMgr.Create(conf); err != nil {
			log.Warnf("Create failed:%s", err.Error())
		}
	case *storagev1.Iscsi:
		conf := e.Object.(*storagev1.Iscsi)
		if err := d.iscsiMgr.Create(conf); err != nil {
			log.Warnf("Create failed:%s", err.Error())
		}
	case *storagev1.Nfs:
		conf := e.Object.(*storagev1.Nfs)
		if err := d.nfsMgr.Create(conf); err != nil {
			log.Warnf("Create failed:%s", err.Error())
		}
	case *k8sstorage.VolumeAttachment:
		if err := d.CreateFinalizer(obj); err != nil {
			log.Warnf("Watch to volumeAttachment create event, but add finalizer failed:%s", err.Error())
		}
	}
	return handler.Result{}, nil
}

func (d *Controller) OnUpdate(e event.UpdateEvent) (handler.Result, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	switch e.ObjectOld.(type) {
	case *storagev1.Cluster:
		oldc := e.ObjectOld.(*storagev1.Cluster)
		newc := e.ObjectNew.(*storagev1.Cluster)
		if !reflect.DeepEqual(oldc.Status, newc.Status) {
			return handler.Result{}, nil
		} else if !reflect.DeepEqual(oldc.Spec, newc.Spec) {
			if err := d.clusterMgr.Update(oldc, newc); err != nil {
				log.Warnf("Update failed:%s", err.Error())
			}
		} else {
			if newc.DeletionTimestamp != nil && slice.SliceIndex(newc.Finalizers, common.StorageInUsedFinalizer) == -1 {
				if err := d.clusterMgr.Delete(newc); err != nil {
					log.Warnf("Delete failed:%s", err.Error())
				}
			}
		}
	case *storagev1.Iscsi:
		oldc := e.ObjectOld.(*storagev1.Iscsi)
		newc := e.ObjectNew.(*storagev1.Iscsi)
		if !reflect.DeepEqual(oldc.Status, newc.Status) {
			return handler.Result{}, nil
		} else if !reflect.DeepEqual(oldc.Spec.Initiators, newc.Spec.Initiators) {
			if err := d.iscsiMgr.Update(oldc, newc); err != nil {
				log.Warnf("Update failed:%s", err.Error())
			}
		} else {
			if newc.DeletionTimestamp != nil && slice.SliceIndex(newc.Finalizers, common.StorageInUsedFinalizer) == -1 {
				if err := d.iscsiMgr.Delete(newc); err != nil {
					log.Warnf("Delete failed:%s", err.Error())
				}
			}
		}
	case *storagev1.Nfs:
		oldc := e.ObjectOld.(*storagev1.Nfs)
		newc := e.ObjectNew.(*storagev1.Nfs)
		if !reflect.DeepEqual(oldc.Status, newc.Status) {
			return handler.Result{}, nil
		} else {
			if newc.DeletionTimestamp != nil && slice.SliceIndex(newc.Finalizers, common.StorageInUsedFinalizer) == -1 {
				if err := d.nfsMgr.Delete(newc); err != nil {
					log.Warnf("Delete failed:%s", err.Error())
				}
			}
		}
	}
	return handler.Result{}, nil
}

func (d *Controller) OnDelete(e event.DeleteEvent) (handler.Result, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	switch obj := e.Object.(type) {
	case *k8sstorage.VolumeAttachment:
		if err := d.DeleteFinalizer(obj); err != nil {
			log.Warnf("Watch to volumeAttachment delete event, but del finalizer failed:%s", err.Error())
		}
	case *storagev1.Cluster:
		if slice.SliceIndex(obj.Finalizers, common.StoragePrestopHookFinalizer) == -1 {
			if err := d.clusterMgr.Delete(obj); err != nil {
				log.Warnf("Delete failed:%s", err.Error())
			}
		}
	case *storagev1.Iscsi:
		conf := e.Object.(*storagev1.Iscsi)
		if slice.SliceIndex(conf.Finalizers, common.StoragePrestopHookFinalizer) == -1 {
			if err := d.iscsiMgr.Delete(conf); err != nil {
				log.Warnf("Delete failed:%s", err.Error())
			}
		}
	case *storagev1.Nfs:
		conf := e.Object.(*storagev1.Nfs)
		if slice.SliceIndex(conf.Finalizers, common.StoragePrestopHookFinalizer) == -1 {
			if err := d.nfsMgr.Delete(conf); err != nil {
				log.Warnf("Delete failed:%s", err.Error())
			}
		}
	}
	return handler.Result{}, nil
}

func (d *Controller) OnGeneric(e event.GenericEvent) (handler.Result, error) {
	return handler.Result{}, nil
}

func (d *Controller) CreateFinalizer(va *k8sstorage.VolumeAttachment) error {
	driver := va.Spec.Attacher
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
func (d *Controller) DeleteFinalizer(va *k8sstorage.VolumeAttachment) error {
	lastone, err := common.IsLastOne(d.client, va)
	if err != nil {
		return err
	}
	if !lastone {
		return nil
	}
	driver := va.Spec.Attacher
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
