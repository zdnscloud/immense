package ceph

import (
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/event"
	"github.com/zdnscloud/gok8s/handler"
	"github.com/zdnscloud/immense/pkg/ceph/global"
	cephhandle "github.com/zdnscloud/immense/pkg/ceph/handle"
	corev1 "k8s.io/api/core/v1"
)

func (s *Ceph) OnCreate(e event.CreateEvent) (handler.Result, error) {
	return handler.Result{}, nil
}
func (s *Ceph) OnUpdate(e event.UpdateEvent) (handler.Result, error) {
	switch newobj := e.ObjectNew.(type) {
	case *corev1.Endpoints:
		if newobj.Name != global.MonSvc {
			return handler.Result{}, nil
		}
		if err := cephhandle.DoEndpointsUpdate(s.cli, e.ObjectOld.(*corev1.Endpoints), newobj); err != nil {
			log.Warnf("Endpoints update handler failed:%s", err.Error())
		}
	}
	return handler.Result{}, nil
}
func (s *Ceph) OnDelete(e event.DeleteEvent) (handler.Result, error) {
	return handler.Result{}, nil
}
func (s *Ceph) OnGeneric(e event.GenericEvent) (handler.Result, error) {
	return handler.Result{}, nil
}
