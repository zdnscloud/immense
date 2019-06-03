package common

import (
	"bytes"
	"context"
	"fmt"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	corev1 "k8s.io/api/core/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"strings"
	"text/template"
)

const (
	RBACConfig               = "rbac"
	StorageHostLabels        = "storage.zcloud.cn/storagetype"
	StorageBlocksAnnotations = "storage.zcloud.cn/blocks"
	StorageNamespace         = "zcloud"
)

func CreateNodeAnnotationsAndLabels(cli client.Client, cluster *storagev1.Cluster, nodelabelvalue string) error {
	for _, host := range cluster.Spec.Hosts {
		log.Debugf("Add Annotations and Labels fot host:%s", host.NodeName)
		node := corev1.Node{}
		if err := cli.Get(context.TODO(), k8stypes.NamespacedName{"", host.NodeName}, &node); err != nil {
			return err
		}
		node.Labels[StorageHostLabels] = nodelabelvalue
		node.Annotations[StorageBlocksAnnotations] = strings.Replace(strings.Trim(fmt.Sprint(host.BlockDevices), "[]"), " ", ",", -1)
		if err := cli.Update(context.TODO(), &node); err != nil {
			return err
		}
	}
	return nil
}

func CompileTemplateFromMap(tmplt string, configMap interface{}) (string, error) {
	out := new(bytes.Buffer)
	t := template.Must(template.New("compiled_template").Parse(tmplt))
	if err := t.Execute(out, configMap); err != nil {
		return "", err
	}
	return out.String(), nil
}

func DeleteNodeAnnotationsAndLabels(cli client.Client, cluster *storagev1.Cluster, nodelabelvalue string) error {
	for _, host := range cluster.Spec.Hosts {
		log.Debugf("Del Annotations and Labels fot host:%s", host.NodeName)
		node := corev1.Node{}
		if err := cli.Get(context.TODO(), k8stypes.NamespacedName{"", host.NodeName}, &node); err != nil {
			return err
		}
		delete(node.Labels, StorageHostLabels)
		delete(node.Annotations, StorageBlocksAnnotations)
	}
	return nil
}

func GetHostAddr(ctx context.Context, cli client.Client, name string) (string, error) {
	node := corev1.Node{}
	if err := cli.Get(ctx, k8stypes.NamespacedName{"", name}, &node); err != nil {
		return "", err
	}
	return node.Annotations["zdnscloud.cn/internal-ip"], nil
}
