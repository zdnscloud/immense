package controller

import (
	"context"
	"fmt"
	"sync"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/cache"
	"github.com/zdnscloud/gok8s/client"
	//k8scfg "github.com/zdnscloud/gok8s/client/config"
	"github.com/zdnscloud/gok8s/controller"
	"github.com/zdnscloud/gok8s/event"
	"github.com/zdnscloud/gok8s/handler"
	"github.com/zdnscloud/gok8s/predicate"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	cephGlobal "github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/common"
	"github.com/zdnscloud/immense/pkg/eventhandler"
	"github.com/zdnscloud/immense/pkg/lvm"
	k8sstorage "k8s.io/api/storage/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

var ctx = context.TODO()

const (
	ClusterInUseFinalizer       = "storage.zcloud.cn/inuse"
	ClusterPrestopHookFinalizer = "storage.zcloud.cn/prestophook"
)

type Controller struct {
	stopCh     chan struct{}
	handlermgr *eventhandler.HandlerManager
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
	/*
		cfg, err := k8scfg.GetConfig()
		if err != nil {
			return nil, err
		}*/
	var options client.Options
	options.Scheme = client.GetDefaultScheme()
	storagev1.AddToScheme(options.Scheme)

	cli, err := client.New(config, options)
	if err != nil {
		return nil, err
	}

	hm := eventhandler.New(cli)

	stopCh := make(chan struct{})
	go c.Start(stopCh)
	c.WaitForCacheSync(stopCh)

	storageCtrl := &Controller{
		stopCh:     stopCh,
		handlermgr: hm,
		client:     cli,
	}
	ctrl := controller.New("zcloudStorage", c, scm)
	ctrl.Watch(&storagev1.Cluster{})
	ctrl.Watch(&k8sstorage.VolumeAttachment{})
	ctrl.Start(stopCh, storageCtrl, predicate.NewIgnoreUnchangedUpdate())
	return storageCtrl, nil
}

func (d *Controller) OnCreate(e event.CreateEvent) (handler.Result, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	switch obj := e.Object.(type) {
	case *storagev1.Cluster:
		cluster := e.Object.(*storagev1.Cluster)
		if err := d.handlermgr.Create(cluster); err != nil {
			log.Warnf("Create failed:%s", err.Error())
		}
	case *k8sstorage.VolumeAttachment:
		d.CreateFinalizer(obj)
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
		if err := d.handlermgr.Update(oldc, newc); err != nil {
			log.Warnf("Update failed:%s", err.Error())
		}
	}
	return handler.Result{}, nil
}

func (d *Controller) OnDelete(e event.DeleteEvent) (handler.Result, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	switch obj := e.Object.(type) {
	case *storagev1.Cluster:
		cluster := e.Object.(*storagev1.Cluster)
		if err := d.handlermgr.Delete(cluster); err != nil {
			log.Warnf("Delete failed:%s", err.Error())
		}
	case *k8sstorage.VolumeAttachment:
		d.DeleteFinalizer(obj)
	}
	return handler.Result{}, nil
}

func (d *Controller) OnGeneric(e event.GenericEvent) (handler.Result, error) {
	return handler.Result{}, nil
}

func (d *Controller) CreateFinalizer(va *k8sstorage.VolumeAttachment) error {
	storageType, err := getStorageTypeFromVolumeAttachment(va)
	if err != nil {
		return err
	}
	return common.AddFinalizerForStorage(d.client, storageType, ClusterInUseFinalizer)
}
func (d *Controller) DeleteFinalizer(va *k8sstorage.VolumeAttachment) error {
	lastone, err := common.IsLastOne(d.client, va)
	if err != nil {
		return err
	}
	if !lastone {
		return nil
	}
	storageType, err := getStorageTypeFromVolumeAttachment(va)
	if err != nil {
		return err
	}
	return common.DelFinalizerForStorage(d.client, storageType, ClusterInUseFinalizer)
}

func getStorageTypeFromVolumeAttachment(va *k8sstorage.VolumeAttachment) (string, error) {
	switch va.Spec.Attacher {
	case lvm.StorageDriverName:
		return lvm.StorageType, nil
	case cephGlobal.StorageDriverName:
		return cephGlobal.StorageType, nil
	default:
		return "", fmt.Errorf("unknow storage cluster type for volumeAttachment %s", va.Name)
	}
}
