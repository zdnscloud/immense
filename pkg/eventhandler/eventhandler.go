package eventhandler

import (
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/ceph"
	"github.com/zdnscloud/immense/pkg/lvm"
)

type Handler interface {
	GetType() string
	Create(cluster *storagev1.Cluster) error
	Delete(cluster *storagev1.Cluster) error
	Update(oldc, newc *storagev1.Cluster) error
}

type HandlerManager struct {
	handlers []Handler
}

func New(cli client.Client) *HandlerManager {
	return &HandlerManager{
		handlers: []Handler{
			lvm.New(cli),
			ceph.New(cli),
		},
	}
}

func (h *HandlerManager) Create(cluster *storagev1.Cluster) error {
	for _, s := range h.handlers {
		if s.GetType() == cluster.Spec.StorageType {
			return s.Create(cluster)
		}
	}
	return nil
}

func (h *HandlerManager) Delete(cluster *storagev1.Cluster) error {
	for _, s := range h.handlers {
		if s.GetType() == cluster.Spec.StorageType {
			return s.Delete(cluster)
		}
	}
	return nil
}

func (h *HandlerManager) Update(oldc *storagev1.Cluster, newc *storagev1.Cluster) error {
	for _, s := range h.handlers {
		if s.GetType() == oldc.Spec.StorageType {
			return s.Update(oldc, newc)
		}
	}
	return nil
}
