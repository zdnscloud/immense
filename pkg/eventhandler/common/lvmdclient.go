package common

import (
	"context"
	lvmdclient "github.com/zdnscloud/lvmd/client"
	pb "github.com/zdnscloud/lvmd/proto"
	"net"
	"time"
)

const (
	VGName = "k8s"
)

func Validate(ctx context.Context, lvmdcli *lvmdclient.Client, block string) (bool, error) {
	req := pb.ValidateRequest{
		Block: block,
	}
	out, err := lvmdcli.Validate(ctx, &req)
	return out.Validate, err
}

func PvExist(ctx context.Context, lvmdcli *lvmdclient.Client, block string) (bool, error) {
	pvsreq := pb.ListPVRequest{}
	pvsout, err := lvmdcli.ListPV(ctx, &pvsreq)
	if err != nil {
		return false, err
	}
	for _, v := range pvsout.Pvinfos {
		if v.Name == block && v.Fmt == "lvm2" {
			return true, nil
		}
	}
	return false, nil
}

func VgExist(ctx context.Context, lvmdcli *lvmdclient.Client) (bool, error) {
	vgsreq := pb.ListVGRequest{}
	vgsout, err := lvmdcli.ListVG(ctx, &vgsreq)
	if err != nil {
		return false, err
	}
	for _, v := range vgsout.VolumeGroups {
		if v.Name == VGName {
			return true, nil
		}
	}
	return false, nil
}

func PvCreate(ctx context.Context, lvmdcli *lvmdclient.Client, block string) error {
	req := pb.CreatePVRequest{
		Block: block,
	}
	_, err := lvmdcli.CreatePV(ctx, &req)
	return err
}

func VgCreate(ctx context.Context, lvmdcli *lvmdclient.Client, block string) error {
	req := pb.CreateVGRequest{
		Name:           VGName,
		PhysicalVolume: block,
	}
	_, err := lvmdcli.CreateVG(ctx, &req)
	return err
}

func VgExtend(ctx context.Context, lvmdcli *lvmdclient.Client, block string) error {
	req := pb.ExtendVGRequest{
		Name:           VGName,
		PhysicalVolume: block,
	}
	_, err := lvmdcli.ExtendVG(ctx, &req)
	return err
}

func WaitLvmd(addr string) bool {
	for i := 0; i < 20; i++ {
		_, err := net.Dial("tcp", addr)
		if err == nil {
			return true
		}
		time.Sleep(6 * time.Second)
	}
	return false
}

func GetVG(ctx context.Context, lvmdcli *lvmdclient.Client, block string) (string, error) {
	req := pb.MatchRequest{
		Block: block,
	}
	out, err := lvmdcli.Match(ctx, &req)
	return out.CommandOutput, err
}

func Destory(ctx context.Context, lvmdcli *lvmdclient.Client, block string) (string, error) {
	req := pb.DestoryRequest{
		Block: block,
	}
	out, err := lvmdcli.Destory(ctx, &req)
	return out.CommandOutput, err
}
