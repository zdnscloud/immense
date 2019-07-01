package main

import (
	"context"
	"fmt"
	k8scli "github.com/zdnscloud/gok8s/client"
	k8scfg "github.com/zdnscloud/gok8s/client/config"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
)

func main() {
	cfg, err := k8scfg.GetConfig()
	if err != nil {
		fmt.Println("1111", err)
	}

	scm := scheme.Scheme
	storagev1.AddToScheme(scm)
	cli, err := k8scli.New(cfg, k8scli.Options{
		Scheme: scm,
	})
	if err != nil {
		fmt.Println("2222", err)
	}
	/*
		storagecluster := storagev1.Cluster{}
		err = cli.Get(context.TODO(), k8stypes.NamespacedName{"default", "example-cluster"}, &storagecluster)
	*/
	/*
		storagecluster := &storagev1.Cluster{
			ObjectMeta: metav1.ObjectMeta{Name: "example-cluster", Namespace: "default"},
		}
		err = cli.Delete(context.TODO(), storagecluster)
	*/
	/*
		storagecluster := &storagev1.Cluster{
			ObjectMeta: metav1.ObjectMeta{Name: "example-cluster", Namespace: "default"},
			Spec: storagev1.ClusterSpec{
				StorageType: "ceph",
				Hosts: []storagev1.HostSpec{
					storagev1.HostSpec{
						NodeName:     "k8s-02",
						BlockDevices: []string{"/dev/vdb"},
					},
					storagev1.HostSpec{
						NodeName:     "k8s-04",
						BlockDevices: []string{"/dev/vdb"},
					},
					storagev1.HostSpec{
						NodeName:     "k8s-06",
						BlockDevices: []string{"/dev/vdb", "/dev/vdc"},
					},
				},
			},
		}
		err = cli.Create(context.TODO(), storagecluster)
	*/
	/*
		storagecluster := storagev1.Cluster{}
		err = cli.Get(context.TODO(), k8stypes.NamespacedName{"default", "example-cluster"}, &storagecluster)
		storagecluster.Spec.Hosts[2] = storagev1.HostSpec{
			NodeName:     "k8s-06",
			BlockDevices: []string{"/dev/vdb"},
		}
		err = cli.Update(context.TODO(), &storagecluster)
	*/
	storagecluster := storagev1.ClusterList{}
	err = cli.List(context.TODO(), &k8scli.ListOptions{Namespace: "default"}, &storagecluster)
	if err != nil {
		fmt.Println("3333", err)
	}
	fmt.Println(storagecluster)
}
