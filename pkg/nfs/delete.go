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

func delete(cli client.Client, conf *storagev1.Nfs) error {
	if err := unDeployNfsCSI(cli, conf); err != nil {
		return err
	}
	if err := unDeployStorageClass(cli, conf); err != nil {
		return err
	}
	return nil
}

func unDeployNfsCSI(cli client.Client, conf *storagev1.Nfs) error {
	log.Debugf("Undeploy nfs %s csi", conf.Name)

	yaml, err := csiDpyaml(conf.Name, conf.Spec.Server, conf.Spec.Path)
	if err != nil {
		return err
	}
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	if err := common.WaitTerminated(common.DeploymentObj(), cli, common.StorageNamespace, fmt.Sprintf("%s-%s", conf.Name, NfsCSIDpSuffix)); err != nil {
		return err
	}
	yaml, err = csiSayaml(conf.Name)
	if err != nil {
		return err
	}
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
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
