package main

import (
	"fmt"
	"github.com/zdnscloud/gok8s/cache"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/client/config"
	"github.com/zdnscloud/gok8s/controller"
	"github.com/zdnscloud/gok8s/event"
	"github.com/zdnscloud/gok8s/handler"
	"github.com/zdnscloud/gok8s/predicate"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sync"
)

type Controller struct {
	stopCh chan struct{}
	lock   sync.RWMutex
	client client.Client
}

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Println(err)
	}
	scm := scheme.Scheme
	storagev1.AddToScheme(scm)

	opts := cache.Options{
		Scheme: scm,
	}
	c, err := cache.New(cfg, opts)
	if err != nil {
		fmt.Println(err)
	}

	var options client.Options
	options.Scheme = client.GetDefaultScheme()
	storagev1.AddToScheme(options.Scheme)

	cli, err := client.New(cfg, options)
	if err != nil {
		fmt.Println(err)
	}
	stopCh := make(chan struct{})
	go c.Start(stopCh)
	c.WaitForCacheSync(stopCh)
	storageCtrl := &Controller{
		stopCh: stopCh,
		client: cli,
	}

	ctrl := controller.New("zcloudStorage", c, scm)
	ctrl.Watch(&storagev1.Cluster{})
	ctrl.Start(stopCh, storageCtrl, predicate.NewIgnoreUnchangedUpdate())
}

func (d *Controller) OnCreate(e event.CreateEvent) (handler.Result, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	switch obj := e.Object.(type) {
	case *storagev1.Cluster:
		cluster := e.Object.(*storagev1.Cluster)
		fmt.Println("!!!!!create", cluster, obj)
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
		fmt.Println("!!!!!update", oldc)
		fmt.Println("!!!!!update", newc)
	}
	return handler.Result{}, nil
}
func (d *Controller) OnDelete(e event.DeleteEvent) (handler.Result, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	switch obj := e.Object.(type) {
	case *storagev1.Cluster:
		cluster := e.Object.(*storagev1.Cluster)
		fmt.Println("!!!!!delete", cluster, obj)
	}
	return handler.Result{}, nil
}
func (d *Controller) OnGeneric(e event.GenericEvent) (handler.Result, error) {
	return handler.Result{}, nil
}
