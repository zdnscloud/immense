package main

import (
	//"context"
	"fmt"
	k8scli "github.com/zdnscloud/gok8s/client"
	k8scfg "github.com/zdnscloud/gok8s/client/config"
	"github.com/zdnscloud/gok8s/helper"
)

const (
	RBACConfig               = "rbac"
	StorageHostLabels        = "storage.zcloud.cn/storagetype"
	StorageBlocksAnnotations = "storage.zcloud.cn/blocks"
	StorageNamespace         = "zcloud"
)

const yaml = `

---
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

func main() {
	cfg, err := k8scfg.GetConfig()
	if err != nil {
		fmt.Println(err)
	}
	cli, err := k8scli.New(cfg, k8scli.Options{})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(helper.CreateResourceFromYaml(cli, yaml))
}
