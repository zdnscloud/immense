package main

import (
	"context"
	"fmt"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/client/config"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	corev1 "k8s.io/api/core/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
)

var ctx = context.TODO()

func getChap(cli client.Client, namespace, name string) (string, string, error) {
	secret := corev1.Secret{}
	err := cli.Get(ctx, k8stypes.NamespacedName{namespace, name}, &secret)
	if err != nil {
		return "", "", err
	}
	var username, password string
	for k, v := range secret.Data {
		if k == "username" {
			username = string(v)
		}
		if k == "password" {
			password = string(v)
		}
	}
	return username, password, nil
}

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Println(err)
	}
	scm := scheme.Scheme
	storagev1.AddToScheme(scm)

	var options client.Options
	options.Scheme = client.GetDefaultScheme()
	storagev1.AddToScheme(options.Scheme)

	cli, err := client.New(cfg, options)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(getChap(cli, "zcloud", "iiiii-1-iscsi-secret"))
}
