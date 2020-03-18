package nfs

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/common"
)

func unDeployNfsCSI(cli client.Client, conf *storagev1.Nfs) error {
	log.Debugf("Undeploy nfs %s csi", conf.Name)

	yaml, err := csiyaml(conf.Name, conf.Spec.Server, conf.Spec.Path)
	if err != nil {
		return err
	}
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	common.WaitDpTerminated(cli, common.StorageNamespace, fmt.Sprintf("%s-%s", conf.Name, NfsCSIDpSuffix))
	return nil
}

func unDeployStorageClass(cli client.Client, conf *storagev1.Nfs) error {
	log.Debugf("Undeploy nfs %s storageclass", conf.Name)
	yaml, err := scyaml(conf.Name)
	if err != nil {
		return err
	}
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	return nil
}
