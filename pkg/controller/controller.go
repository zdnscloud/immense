package controller

import (
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/cache"
	"github.com/zdnscloud/gok8s/client"
	k8scfg "github.com/zdnscloud/gok8s/client/config"
	"github.com/zdnscloud/gok8s/controller"
	"github.com/zdnscloud/gok8s/event"
	"github.com/zdnscloud/gok8s/handler"
	"github.com/zdnscloud/gok8s/predicate"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/eventhandler"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type Controller struct {
	stopCh     chan struct{}
	handlermgr *eventhandler.HandlerManager
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

	cfg, err := k8scfg.GetConfig()
	if err != nil {
		return nil, err
	}
	var options client.Options
	options.Scheme = client.GetDefaultScheme()
	storagev1.AddToScheme(options.Scheme)

	cli, err := client.New(cfg, options)
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
	}
	ctrl := controller.New("zcloudStorage", c, scm)
	ctrl.Watch(&storagev1.Cluster{})
	ctrl.Start(stopCh, storageCtrl, predicate.NewIgnoreUnchangedUpdate())
	return storageCtrl, nil
}

func logCluster(cluster *storagev1.Cluster) {
	log.Debugf("name:%s, type:%s", cluster.Name, cluster.Spec.StorageType)
	for _, host := range cluster.Spec.Hosts {
		log.Debugf("node:%s, devices:%s", host.NodeName, host.BlockDevices)
	}
}

func (d *Controller) OnCreate(e event.CreateEvent) (handler.Result, error) {
	log.Debugf("create event")
	cluster := e.Object.(*storagev1.Cluster)
	logCluster(cluster)
	if err := d.handlermgr.Create(cluster); err != nil {
		log.Warnf("Create failed:%s", err.Error())
	}
	return handler.Result{}, nil
}

func (d *Controller) OnUpdate(e event.UpdateEvent) (handler.Result, error) {
	log.Debugf("update event")
	oldc := e.ObjectOld.(*storagev1.Cluster)
	newc := e.ObjectNew.(*storagev1.Cluster)
	if err := d.handlermgr.Update(oldc, newc); err != nil {
		log.Warnf("Update failed:%s", err.Error())
	}
	return handler.Result{}, nil
}

func (d *Controller) OnDelete(e event.DeleteEvent) (handler.Result, error) {
	log.Debugf("delete event")
	cluster := e.Object.(*storagev1.Cluster)
	logCluster(cluster)
	if err := d.handlermgr.Delete(cluster); err != nil {
		log.Warnf("Delete failed:%s", err.Error())
	}
	return handler.Result{}, nil
}

func (d *Controller) OnGeneric(e event.GenericEvent) (handler.Result, error) {
	return handler.Result{}, nil
}
