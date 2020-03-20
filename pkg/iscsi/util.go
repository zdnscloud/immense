package iscsi

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	k8stypes "k8s.io/apimachinery/pkg/types"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/gok8s/helper"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"github.com/zdnscloud/immense/pkg/common"
	nodeagaentclient "github.com/zdnscloud/node-agent/client"
	pb "github.com/zdnscloud/node-agent/proto"
)

func getIscsi(cli client.Client, name string) (*storagev1.Iscsi, error) {
	iscsi := &storagev1.Iscsi{}
	if err := cli.Get(ctx, k8stypes.NamespacedName{"", name}, iscsi); err != nil {
		return nil, err
	}
	return iscsi, nil
}

func updateStatus(cli client.Client, name string, status storagev1.IscsiStatus) error {
	iscsi, err := getIscsi(cli, name)
	if err != nil {
		return err
	}
	iscsi.Status = status
	return cli.Update(ctx, iscsi)
}

func AddFinalizer(cli client.Client, name, finalizer string) error {
	iscsi, err := getIscsi(cli, name)
	if err != nil {
		return err
	}
	helper.AddFinalizer(iscsi, finalizer)
	log.Debugf("Add finalizer %s for storage iscsi %s", finalizer, name)
	return cli.Update(ctx, iscsi)
}

func RemoveFinalizer(cli client.Client, name, finalizer string) error {
	iscsi, err := getIscsi(cli, name)
	if err != nil {
		return err
	}
	helper.RemoveFinalizer(iscsi, finalizer)
	log.Debugf("Delete finalizer %s for storage iscsi %s", finalizer, name)
	return cli.Update(ctx, iscsi)
}

func UpdateStatusPhase(cli client.Client, name string, phase storagev1.StatusPhase) {
	iscsi, err := getIscsi(cli, name)
	if err != nil {
		if apierrors.IsNotFound(err) == true {
			return
		}
		log.Warnf("Update storage iscsi %s status failed. Err: %s", name, err.Error())
		return
	}
	iscsi.Status.Phase = phase
	if err := cli.Update(ctx, iscsi); err != nil {
		if apierrors.IsNotFound(err) == true {
			return
		}
		log.Warnf("Update storage iscsi %s status failed. Err: %s", name, err.Error())
		return
	}
	return
}

func loginIscsi(cli pb.NodeAgentClient, host, port, iqn, username, password string) error {
	_, err := cli.IscsiDiscovery(ctx, &pb.IscsiDiscoveryRequest{
		Host: host,
		Port: port,
		Iqn:  iqn,
	})
	if err != nil {
		return fmt.Errorf("iscsi discovery failed. %v", err)
	}
	if username != "" && password != "" {
		_, err := cli.IscsiChap(ctx, &pb.IscsiChapRequest{
			Host:     host,
			Port:     port,
			Iqn:      iqn,
			Username: username,
			Password: password,
		})
		if err != nil {
			return fmt.Errorf("iscsi chap failed. %v", err)
		}
	}
	_, err = cli.IscsiLogin(ctx, &pb.IscsiLoginRequest{
		Host: host,
		Port: port,
		Iqn:  iqn,
	})
	if err != nil {
		return fmt.Errorf("iscsi login failed. %v", err)
	}
	reply, err := cli.IsTargetLoggedIn(ctx, &pb.IsTargetLoggedInRequest{
		Host: host,
		Port: port,
		Iqn:  iqn,
	})
	if err != nil {
		return fmt.Errorf("iscsi login check failed. %v", err)
	}
	if !reply.Login {
		return fmt.Errorf("can not find target session")
	}
	return nil
}

func cleanIscsi(cli pb.NodeAgentClient, device string) error {
	if _, err := cli.CleanIscsiDevice(ctx, &pb.CleanIscsiDeviceRequest{
		Device: device,
	}); err != nil {
		return fmt.Errorf("iscsi clean failed. %v", err)
	}
	return nil
}

func reloadMultipath(cli pb.NodeAgentClient) error {
	_, err := cli.ReloadMultipath(ctx, &pb.ReloadMultipathRequest{})
	return err
}

func logoutIscsi(cli pb.NodeAgentClient, host, port, iqn string) error {
	reply, err := cli.IsTargetLoggedIn(ctx, &pb.IsTargetLoggedInRequest{
		Host: host,
		Port: port,
		Iqn:  iqn,
	})
	if err != nil {
		return fmt.Errorf("iscsi login check failed. %v", err)
	}
	if !reply.Login {
		return nil
	}
	if _, err := cli.IscsiLogout(ctx, &pb.IscsiLogoutRequest{
		Host: host,
		Port: port,
		Iqn:  iqn,
	}); err != nil {
		return fmt.Errorf("iscsi login failed. %v", err)
	}
	return nil
}

func getIscsiDevices(cli pb.NodeAgentClient, iqn string) ([]string, error) {
	blocks, err := cli.IscsiGetBlocks(ctx, &pb.IscsiGetBlocksRequest{
		Iqn: iqn,
	})
	if err != nil {
		return nil, fmt.Errorf("iscsi get devices failed. %v", err)
	}
	var devices []string
	for _, info := range blocks.IscsiBlock {
		if len(info.Blocks) > 1 {
			path, err := cli.IscsiGetMultipaths(ctx, &pb.IscsiGetMultipathsRequest{
				Devs: info.Blocks,
			})
			if err != nil {
				return nil, fmt.Errorf("iscsi get multipath failed. %v", err)
			}
			devices = append(devices, multipathDir+path.Dev)
		} else {
			for _, dev := range info.Blocks {
				devices = append(devices, deviceDir+dev)
			}
		}
	}
	if len(devices) == 0 {
		return nil, fmt.Errorf("can not find devices for target %s", iqn)
	}
	return devices, nil
}

func createNodeAgentClient(cli client.Client, node string) (pb.NodeAgentClient, error) {
	addr, err := common.GetHostAddr(ctx, cli, node)
	if err != nil {
		return nil, err
	}
	return nodeagaentclient.NewClient(addr+":"+NodeAgentPort, 10*time.Second)
}

func replaceInitiatorname(cli pb.NodeAgentClient, src, dst string) error {
	if _, err := cli.ReplaceInitiatorname(ctx, &pb.ReplaceInitiatornameRequest{
		SrcFile: src,
		DstFile: dst}); err != nil {
		return fmt.Errorf("replace initiatorname failed. %v", err)
	}
	return nil
}

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
