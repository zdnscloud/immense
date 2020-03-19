package main

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/client/config"

	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
)

func main() {
	config, err := config.GetConfig()
	if err != nil {
		fmt.Println("111 err:", err)
	}
	var options client.Options
	options.Scheme = client.GetDefaultScheme()
	storagev1.AddToScheme(options.Scheme)

	cli, err := client.New(config, options)
	if err != nil {
		fmt.Println("222 err:", err)
	}
	//{[202.173.9.12 202.173.9.14] 3260  iqn.2020-03.zdns.cn:lun1 true [slave]}
	ts := []string{"202.173.9.12", "202.173.9.14"}
	ns := []string{"202.173.9.12", "202.173.9.14"}
	k8sIscsi := &storagev1.Iscsi{
		ObjectMeta: metav1.ObjectMeta{
			Name: "haha",
		},
		Spec: storagev1.IscsiSpec{
			//		Targets:    ["202.173.9.12","202.173.9.14"],
			Targets:    ts,
			Port:       "3260",
			Iqn:        "iqn.2020-03.zdns.cn:lun2",
			Chap:       false,
			Initiators: ns,
		},
	}
	fmt.Println("2222", k8sIscsi.Spec)
	err = cli.Create(context.TODO(), k8sIscsi)
	if err != nil {
		fmt.Println("3333", err)
	}
}
