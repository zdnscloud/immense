package lvm

import (
	"context"
	"errors"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/immense/pkg/common"
	lvmdclient "github.com/zdnscloud/lvmd/client"
	pb "github.com/zdnscloud/lvmd/proto"
	corev1 "k8s.io/api/core/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"net"
	"strconv"
	"time"
)

func getHostAddr(ctx context.Context, cli client.Client, name string) (string, error) {
	node := corev1.Node{}
	if err := cli.Get(ctx, k8stypes.NamespacedName{"", name}, &node); err != nil {
		return "", err
	}
	return node.Annotations[common.NodeIPLabels], nil
}

func CreateLvmdClient(ctx context.Context, cli client.Client, hostname string) (*lvmdclient.Client, error) {
	hostip, err := getHostAddr(ctx, cli, hostname)
	if err != nil {
		return nil, errors.New("Get host address failed!" + err.Error())
	}
	addr := hostip + ":1736"

	if !waitLvmd(addr) {
		return nil, errors.New("Lvmd not ready!" + addr)
	}
	return lvmdclient.New(addr, 5*time.Second)
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

func validate(ctx context.Context, lvmdcli *lvmdclient.Client, block string) (bool, error) {
	req := pb.ValidateRequest{
		Block: block,
	}
	out, err := lvmdcli.Validate(ctx, &req)
	if err != nil {
		return false, err
	}
	return out.Validate, nil
}

func pvExist(ctx context.Context, lvmdcli *lvmdclient.Client, block string) (bool, error) {
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

func vgExist(ctx context.Context, lvmdcli *lvmdclient.Client, name string) (bool, error) {
	vgsreq := pb.ListVGRequest{}
	vgsout, err := lvmdcli.ListVG(ctx, &vgsreq)
	if err != nil {
		return false, err
	}
	for _, v := range vgsout.VolumeGroups {
		if v.Name == name {
			return true, nil
		}
	}
	return false, nil
}

func pvCreate(ctx context.Context, lvmdcli *lvmdclient.Client, block string) (string, error) {
	req := pb.CreatePVRequest{
		Block: block,
	}
	out, err := lvmdcli.CreatePV(ctx, &req)
	if err != nil {
		return "", err
	}
	return out.CommandOutput, nil
}

func vgCreate(ctx context.Context, lvmdcli *lvmdclient.Client, block string, name string) (string, error) {
	req := pb.CreateVGRequest{
		Name:           name,
		PhysicalVolume: block,
	}
	out, err := lvmdcli.CreateVG(ctx, &req)
	if err != nil {
		return "", err
	}
	return out.CommandOutput, nil
}

func vgExtend(ctx context.Context, lvmdcli *lvmdclient.Client, block string, name string) (string, error) {
	req := pb.ExtendVGRequest{
		Name:           name,
		PhysicalVolume: block,
	}
	out, err := lvmdcli.ExtendVG(ctx, &req)
	if err != nil {
		return "", err
	}
	return out.CommandOutput, nil
}

func vgReduce(ctx context.Context, lvmdcli *lvmdclient.Client, block string, name string) (string, error) {
	req := pb.ExtendVGRequest{
		Name:           name,
		PhysicalVolume: block,
	}
	out, err := lvmdcli.ReduceVG(ctx, &req)
	if err != nil {
		return "", err
	}
	return out.CommandOutput, nil
}

func destory(ctx context.Context, lvmdcli *lvmdclient.Client, block string) (string, error) {
	req := pb.DestoryRequest{
		Block: block,
	}
	out, err := lvmdcli.Destory(ctx, &req)
	if err != nil {
		return "", err
	}
	return out.CommandOutput, nil
}

func removeVG(ctx context.Context, lvmdcli *lvmdclient.Client, name string) (string, error) {
	req := pb.CreateVGRequest{
		Name: name,
	}
	out, err := lvmdcli.RemoveVG(ctx, &req)
	if err != nil {
		return "", err
	}
	return out.CommandOutput, nil
}

func removePV(ctx context.Context, lvmdcli *lvmdclient.Client, block string) (string, error) {
	req := pb.RemovePVRequest{
		Block: block,
	}
	out, err := lvmdcli.RemovePV(ctx, &req)
	if err != nil {
		return "", err
	}
	return out.CommandOutput, nil
}

func GetVG(ctx context.Context, lvmdcli *lvmdclient.Client, block string) (string, error) {
	req := pb.MatchRequest{
		Block: block,
	}
	out, err := lvmdcli.Match(ctx, &req)
	if err != nil {
		return "", err
	}
	return out.CommandOutput, nil
}

func getPVNum(ctx context.Context, lvmdcli *lvmdclient.Client, name string) (int, error) {
	req := pb.CreateVGRequest{
		Name: name,
	}
	out, err := lvmdcli.GetPVNum(ctx, &req)
	if err != nil {
		return 0, err
	}
	num, err := strconv.Atoi(out.CommandOutput)
	if err != nil {
		return 0, err
	}
	return num, nil
}

func CreateVG(ctx context.Context, lvmdcli *lvmdclient.Client, block string, name string) error {
	v, err := vgExist(ctx, lvmdcli, name)
	if err != nil {
		return errors.New("Check vg exist failed!" + err.Error())
	}
	if v {
		_, err := vgExtend(ctx, lvmdcli, block, name)
		if err != nil {
			return errors.New("Extend vg failed!" + err.Error())
		}
	} else {
		_, err := vgCreate(ctx, lvmdcli, block, name)
		if err != nil {
			return errors.New("Create vg failed!" + err.Error())
		}
	}
	return nil
}

func RemoveVG(ctx context.Context, lvmdcli *lvmdclient.Client, name string) error {
	v, err := vgExist(ctx, lvmdcli, name)
	if err != nil {
		return errors.New("Check vg exist failed!" + err.Error())
	}
	if v {
		_, err := removeVG(ctx, lvmdcli, name)
		if err != nil {
			return errors.New("Remove vg failed!" + err.Error())
		}
	}
	return nil
}

func CreatePV(ctx context.Context, lvmdcli *lvmdclient.Client, block string) error {
	p, err := pvExist(ctx, lvmdcli, block)
	if err != nil {
		return errors.New("Check pv exist failed!" + err.Error())
	}
	if !p {
		if _, err := pvCreate(ctx, lvmdcli, block); err != nil {
			return errors.New("Create pv exist failed!" + err.Error())
		}
	}
	return nil
}

func RemovePV(ctx context.Context, lvmdcli *lvmdclient.Client, block string) error {
	p, err := pvExist(ctx, lvmdcli, block)
	if err != nil {
		return errors.New("Check pv exist failed!" + err.Error())
		return err
	}
	if p {
		_, err := removePV(ctx, lvmdcli, block)
		if err != nil {
			return errors.New("Remove pv failed!" + err.Error())
		}
	}
	_, err = destory(ctx, lvmdcli, block)
	if err != nil {
		return errors.New("Destory block failed!" + err.Error())
	}
	return nil
}

func VgReduce(ctx context.Context, lvmdcli *lvmdclient.Client, block string, name string) error {
	v, err := vgExist(ctx, lvmdcli, name)
	if err != nil {
		return errors.New("Check vg failed!" + err.Error())
	}
	if v {
		num, err := getPVNum(ctx, lvmdcli, name)
		if err != nil {
			return errors.New("Get vg's pv num failed!" + err.Error())
		}
		if num == 1 {
			_, err := removeVG(ctx, lvmdcli, name)
			if err != nil {
				return errors.New("Remove vg failed!" + err.Error())
			}
		} else {
			_, err := vgReduce(ctx, lvmdcli, block, name)
			if err != nil {
				return errors.New("Reduce vg failed!" + err.Error())
			}
		}
	}
	return RemovePV(ctx, lvmdcli, block)
}

func Validate(ctx context.Context, lvmdcli *lvmdclient.Client, block string) error {
	v, err := validate(ctx, lvmdcli, block)
	if err != nil {
		return errors.New("Validate block failed!" + err.Error())
	}
	if !v {
		_, err := destory(ctx, lvmdcli, block)
		if err != nil {
			return errors.New("Destory block failed!" + err.Error())
		}
	}
	return nil
}