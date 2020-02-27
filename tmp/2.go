package main

import (
	"context"
	"fmt"
	"github.com/zdnscloud/immense/pkg/lvm"
	lvmdclient "github.com/zdnscloud/lvmd/client"
	pb "github.com/zdnscloud/lvmd/proto"

	"time"
)

const (
	addr           = "10.0.0.147:1736"
	addr2          = "10.0.0.148:1736"
	ConnectTimeout = 3 * time.Second
	block          = "/dev/sda"
	VGName         = "iscsi-group"
	pool           = "iscsi-pool"
)

var ctx = context.TODO()

func main() {
	lvmdcli, err := lvmdclient.New(addr, ConnectTimeout)
	if err != nil {
		fmt.Println("11111", err)
	}
	defer lvmdcli.Close()
	//createVG(lvmdcli)
	//creatrPool(lvmdcli)
	//createThinLV(lvmdcli)
	name := "pvc-2020-04"
	createLV(lvmdcli, name)
	return
	lvmdcli2, err := lvmdclient.New(addr2, ConnectTimeout)
	if err != nil {
		fmt.Println("11111", err)
	}
	defer lvmdcli.Close()
	changeLV(lvmdcli2, name)
	//deleteLV(lvmdcli)
	//expandLV(lvmdcli)
}

func createLV(lvmdcli *lvmdclient.Client, name string) {
	resp, err := lvmdcli.CreateLV(ctx, &pb.CreateLVRequest{
		VolumeGroup: VGName,
		Name:        name,
		Size:        2048000,
	})
	if err != nil {
		fmt.Println("99999", err)
		return
	}
	fmt.Println(resp)
}

func changeLV(lvmdcli *lvmdclient.Client, name string) {
	resp, err := lvmdcli.ChangeLV(ctx, &pb.ChangeLVRequest{
		VolumeGroup: VGName,
		Name:        name,
	})
	if err != nil {
		fmt.Println("99999", err)
		return
	}
	fmt.Println(resp)
}
func expandLV(lvmdcli *lvmdclient.Client) {
	resp, err := lvmdcli.ResizeLV(ctx, &pb.ResizeLVRequest{
		VolumeGroup: VGName,
		Name:        "pvc-test06",
		Size:        2048000000,
	})
	if err != nil {
		fmt.Println("99999", err)
		return
	}
	fmt.Println(resp)
}

func deleteLV(lvmdcli *lvmdclient.Client) {
	resp, err := lvmdcli.RemoveLV(ctx, &pb.RemoveLVRequest{
		VolumeGroup: VGName,
		Name:        "pvc-test05",
	})
	if err != nil {
		fmt.Println("77777", err)
		return
	}
	fmt.Println(resp)
}

func createThinLV(lvmdcli *lvmdclient.Client) {
	resp, err := lvmdcli.CreateThinLV(ctx, &pb.CreateThinLVRequest{
		VolumeGroup: VGName,
		Pool:        pool,
		Name:        "pvc-test11",
		Size:        1024000000,
	})
	if err != nil {
		fmt.Println("77777", err)
		return
	}
	fmt.Println(resp)
}

func creatrPool(lvmdcli *lvmdclient.Client) {
	lvsResp, err := lvmdcli.ListLV(ctx, &pb.ListLVRequest{
		VolumeGroup: VGName,
	})
	if err != nil {
		fmt.Println("77777", err)
		return
	}
	for _, l := range lvsResp.Volumes {
		fmt.Println(l.Name)
		if l.Name == pool {
			fmt.Println("pool has already create, skip")
			return
		}
	}
	resp, err := lvmdcli.CreateThinPool(ctx, &pb.CreateThinPoolRequest{
		VolumeGroup: VGName,
		Pool:        pool,
	})
	if err != nil {
		fmt.Println("66666", err)
		return
	}
	fmt.Println("creatrPool", resp)
}

func createVG(lvmdcli *lvmdclient.Client) {
	name, err := lvm.GetVG(ctx, lvmdcli, block)
	if err != nil {
		fmt.Println("22222", err)
		return
	}
	fmt.Println(name)
	if name == VGName {
		fmt.Sprintf("Block had inited before, skip %s\n", block)
		return
	}
	if err := lvm.Validate(ctx, lvmdcli, block); err != nil {
		fmt.Println("33333", err)
		return
	}
	if err := lvm.CreatePV(ctx, lvmdcli, block); err != nil {
		fmt.Println("44444", err)
		return
	}
	if err := lvm.CreateVG(ctx, lvmdcli, block, VGName); err != nil {
		fmt.Println("55555", err)
		return
	}
}
