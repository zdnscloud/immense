package ceph

import (
	"github.com/zdnscloud/gok8s/cache"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/client/config"
	"github.com/zdnscloud/gok8s/controller"
	"github.com/zdnscloud/gok8s/predicate"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

type Ceph struct {
	cli    client.Client
	stopCh chan struct{}
}

func New(c client.Client) (*Ceph, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	ca, err := cache.New(cfg, cache.Options{})
	if err != nil {
		return nil, err
	}
	stopCh := make(chan struct{})
	go ca.Start(stopCh)
	ca.WaitForCacheSync(stopCh)
	cephCtrl := &Ceph{
		cli: c,
	}
	ctrl := controller.New("ceph", ca, scheme.Scheme)
	ctrl.Watch(&corev1.Endpoints{})
	go ctrl.Start(stopCh, cephCtrl, predicate.NewIgnoreUnchangedUpdate())

	return cephCtrl, nil
}

func (s *Ceph) GetType() string {
	return "ceph"
}

func (s *Ceph) Create(cluster *storagev1.Cluster) error {
	if err := common.CreateNodeAnnotationsAndLabels(s.cli, cluster); err != nil {
		return err
	}
	return create(s.cli, cluster)
}

func (s *Ceph) Update(oldcfg, newcfg *storagev1.Cluster) error {
	delcfg, addcfg, changetodel, changetoadd := common.Diff(oldcfg, newcfg)
	if err := doAddhost(s.cli, addcfg); err != nil {
		return err
	}
	if err := doChangeAdd(s.cli, changetoadd); err != nil {
		return err
	}
	if err := doDelhost(s.cli, delcfg); err != nil {
		return err
	}
	if err := doChangeDel(s.cli, changetodel); err != nil {
		return err
	}
	if err := common.DeleteNodeAnnotationsAndLabels(s.cli, oldcfg); err != nil {
		return err
	}
	return common.CreateNodeAnnotationsAndLabels(s.cli, newcfg)
}

func (s *Ceph) Delete(cluster *storagev1.Cluster) error {
	if err := delete(s.cli, cluster); err != nil {
		return err
	}
	return common.DeleteNodeAnnotationsAndLabels(s.cli, cluster)
}
