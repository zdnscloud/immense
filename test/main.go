package main

import (
	"context"
	"fmt"
	k8scli "github.com/zdnscloud/gok8s/client"
	k8scfg "github.com/zdnscloud/gok8s/client/config"
	//"github.com/zdnscloud/gok8s/helper"
	corev1 "k8s.io/api/core/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

const (
	RBACConfig               = "rbac"
	StorageHostLabels        = "storage.zcloud.cn/storagetype"
	StorageBlocksAnnotations = "storage.zcloud.cn/blocks"
	StorageNamespace         = "zcloud"
)

const yaml3 = `
apiVersion: v1
kind: Namespace
metadata:
  name: zdns1
---
apiVersion: v1
kind: Namespace
metadata:
  name: zdns2
`

func del(n *corev1.Node) {
	delete(*n.Labels, StorageHostLabels)
	return
}

func main() {
	cfg, err := k8scfg.GetConfig()
	if err != nil {
		fmt.Println(err)
	}
	cli, err := k8scli.New(cfg, k8scli.Options{})
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(helper.DeleteResourceFromYaml(cli, yaml3))
	node := corev1.Node{}
	if err := cli.Get(context.TODO(), k8stypes.NamespacedName{"", "k8s-02"}, &node); err != nil {
		fmt.Println(err)
	}
	del(node)
	//node.Labels[StorageHostLabels] = "ycy"
	//node.Annotations[StorageBlocksAnnotations] = "test"
	//delete(&node.Labels, StorageHostLabels)
	//delete(&node.Annotations, StorageBlocksAnnotations)
}
