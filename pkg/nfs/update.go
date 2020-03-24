package nfs

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/common"
)

func update(cli client.Client, oldConf, newConf *storagev1.Nfs) error {
	if err := uMountTmpdir(oldConf.Name); err != nil {
		return err
	}
	yaml, err := csiDpyaml(oldConf.Name, oldConf.Spec.Server, oldConf.Spec.Path)
	if err != nil {
		return err
	}
	if err := helper.DeleteResourceFromYaml(cli, yaml); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	if err := common.WaitTerminated(common.DeploymentObj(), cli, common.StorageNamespace, fmt.Sprintf("%s-%s", oldConf.Name, NfsCSIDpSuffix)); err != nil {
		return err
	}
	return create(cli, newConf)
}
