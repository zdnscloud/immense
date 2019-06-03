package lvm

import (
	"context"
	"errors"
	"fmt"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/eventhandler/common"
	lvmdclient "github.com/zdnscloud/lvmd/client"
	pb "github.com/zdnscloud/lvmd/proto"
	corev1 "k8s.io/api/core/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"net"
	"time"
)

const (
	NodeLabelValue = "Lvm"
	LvmdPort       = "1736"
)

var ConLvmdTimeout = 5 * time.Second

func Create(cli client.Client, cluster *storagev1.Cluster) error {
	if err := common.CreateNodeAnnotationsAndLabels(cli, cluster, NodeLabelValue); err != nil {
		return err
	}
	if err := deployLvmd(cli, cluster); err != nil {
		return err
	}
	if err := initBlocks(cli, cluster); err != nil {
		return err
	}
	return nil
	return deployLvmCSI(cli, cluster)
}

func deployLvmCSI(cli client.Client, cluster *storagev1.Cluster) error {
	fmt.Println("deploy for %s", cluster.Spec.StorageType)
	cfg := map[string]interface{}{
		"AoDNamespace":                   "yes",
		"RBACConfig":                     common.RBACConfig,
		"LabelKey":                       common.StorageHostLabels,
		"LabelValue":                     NodeLabelValue,
		"StorageNamespace":               common.StorageNamespace,
		"StorageLvmAttacherImage":        "quay.io/k8scsi/csi-attacher:v1.0.0",
		"StorageLvmProvisionerImage":     "quay.io/k8scsi/csi-provisioner:v1.0.0",
		"StorageLvmDriverRegistrarImage": "quay.io/k8scsi/csi-node-driver-registrar:v1.0.2",
		"StorageLvmCSIImage":             "zdnscloud/lvmcsi:v0.5",
		"StorageClassName":               "lvm",
	}
	yaml, err := common.CompileTemplateFromMap(LvmCSITemplate, cfg)
	if err != nil {
		return err
	}
	return helper.CreateResourceFromYaml(cli, yaml)
}

func deployLvmd(cli client.Client, cluster *storagev1.Cluster) error {
	fmt.Println("deplot lvmd for")
	cfg := map[string]interface{}{
		"AoDNamespace":     "yes",
		"RBACConfig":       common.RBACConfig,
		"LabelKey":         common.StorageHostLabels,
		"LabelValue":       NodeLabelValue,
		"StorageNamespace": common.StorageNamespace,
		"StorageLvmdImage": "zdnscloud/lvmd:v0.92",
	}
	yaml, err := common.CompileTemplateFromMap(LvmDTemplate, cfg)
	if err != nil {
		return err
	}
	return helper.CreateResourceFromYaml(cli, yaml)
}

func initBlocks(cli client.Client, cluster *storagev1.Cluster) error {
	for _, host := range cluster.Spec.Hosts {
		hostip, err := getHostAddr(cli, host.NodeName)
		if err != nil {
			return err
		}
		addr := hostip + ":" + LvmdPort
		if !waitLvmd(addr) {
			return errors.New("Lvmd not ready!" + addr)
		}
		lvmdcli, err := lvmdclient.New(addr, ConLvmdTimeout)
		defer lvmdcli.Close()
		if err != nil {
			return err
		}

		for _, block := range host.BlockDevices {
			fmt.Println("check block:", block)
			v, err := validate(lvmdcli, block)
			if err != nil {
				return err
			}
			if !v {
				return errors.New("some blocks cat not be used!" + block)
			}
		}

		for _, block := range host.BlockDevices {
			fmt.Println("pv :", block)
			p, err := pvExist(lvmdcli, block)
			if err != nil {
				return err
			}
			if p {
				continue
			}
			if err := pvCreate(lvmdcli, block); err != nil {
				return err
			}
		}
		for _, block := range host.BlockDevices {
			fmt.Println("vg :", block)
			v, err := vgExist(lvmdcli)
			if err != nil {
				return err
			}
			if v {
				if vgExtend(lvmdcli, block); err != nil {
					return err
				}
				continue
			}
			if err := vgCreate(lvmdcli, block); err != nil {
				return err
			}
		}
	}
	return nil
}

func getHostAddr(cli client.Client, name string) (string, error) {
	node := corev1.Node{}
	if err := cli.Get(context.TODO(), k8stypes.NamespacedName{"", name}, &node); err != nil {
		return "", err
	}
	return node.Annotations["zdnscloud.cn/internal-ip"], nil
}

func validate(lvmdcli *lvmdclient.Client, block string) (bool, error) {
	req := pb.ValidateRequest{
		Block: block,
	}
	out, err := lvmdcli.Validate(context.TODO(), &req)
	return out.Validate, err
}

func pvExist(lvmdcli *lvmdclient.Client, block string) (bool, error) {
	pvsreq := pb.ListPVRequest{}
	pvsout, err := lvmdcli.ListPV(context.TODO(), &pvsreq)
	if err != nil {
		return false, err
	}
	for _, v := range pvsout.Pvinfos {
		if v.Name == block {
			return true, nil
		}
	}
	return false, nil
}

func vgExist(lvmdcli *lvmdclient.Client) (bool, error) {
	vgsreq := pb.ListVGRequest{}
	vgsout, err := lvmdcli.ListVG(context.TODO(), &vgsreq)
	if err != nil {
		return false, err
	}
	for _, v := range vgsout.VolumeGroups {
		if v.Name == "k8s" {
			return true, nil
		}
	}
	return false, nil
}

func pvCreate(lvmdcli *lvmdclient.Client, block string) error {
	req := pb.CreatePVRequest{
		Block: "/dev/vdb",
	}
	_, err := lvmdcli.CreatePV(context.TODO(), &req)
	return err
}

func vgCreate(lvmdcli *lvmdclient.Client, block string) error {
	req := pb.CreateVGRequest{
		Name:           "k8s",
		PhysicalVolume: block,
	}
	_, err := lvmdcli.CreateVG(context.TODO(), &req)
	return err
}

func vgExtend(lvmdcli *lvmdclient.Client, block string) error {
	req := pb.ExtendVGRequest{
		Name:           "k8s",
		PhysicalVolume: block,
	}
	_, err := lvmdcli.ExtendVG(context.TODO(), &req)
	return err
}

func waitLvmd(addr string) bool {
	for i := 0; i < 20; i++ {
		_, err := net.Dial("tcp", addr)
		if err == nil {
			return true
		}
		time.Sleep(6 * time.Second)
	}
	return false
}
