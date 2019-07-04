package ceph

import (
	"context"
	"errors"
	"github.com/zdnscloud/gok8s/client"
	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/ceph/util"
	"github.com/zdnscloud/immense/pkg/common"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"os"
	"path"
	"strings"
)

var root = "/etc/ceph"
var files = []string{"ceph.conf", "ceph.client.admin.keyring", "ceph.mon.keyring"}

func Rmmon(name string) error {
	args := []string{"mon", "rm", name, "--connect-timeout", "15"}
	_, err := util.ExecCMDWithOutput("ceph", args)
	if err != nil {
		return err
	}
	return nil
}

func ReweigtOsd(id string) error {
	args := []string{"osd", "crush", "reweight", id, "0", "--connect-timeout", "15"}
	_, err := util.ExecCMDWithOutput("ceph", args)
	if err != nil {
		return err
	}
	return nil
}

func OutOsd(id string) error {
	args := []string{"osd", "out", id, "--connect-timeout", "15"}
	_, err := util.ExecCMDWithOutput("ceph", args)
	if err != nil {
		return err
	}
	return nil
}

func RemoveCrush(id string) error {
	args := []string{"osd", "crush", "remove", id, "--connect-timeout", "15"}
	_, err := util.ExecCMDWithOutput("ceph", args)
	if err != nil {
		return err
	}
	return nil
}

func RmOsd(id string) error {
	args := []string{"osd", "rm", id, "--connect-timeout", "15"}
	_, err := util.ExecCMDWithOutput("ceph", args)
	if err != nil {
		return err
	}
	return nil
}

func RmOsdAuth(id string) error {
	args := []string{"auth", "del", id, "--connect-timeout", "15"}
	_, err := util.ExecCMDWithOutput("ceph", args)
	if err != nil {
		return err
	}
	return nil
}

func GetDownOsdIDs(stat string) ([]string, error) {
	ids := make([]string, 0)
	args := []string{"osd", "dump", "--connect-timeout", "15"}
	out, err := util.ExecCMDWithOutput("ceph", args)
	if err != nil {
		return ids, err
	}
	lines := strings.Split(string(out), "\n")
	for _, l := range lines {
		tmp := strings.Fields(l)
		if strings.HasPrefix(tmp[0], "osd") && tmp[1] == "down" && !strings.Contains(tmp[len(tmp)-2], "new") && tmp[2] == stat {
			ids = append(ids, strings.Split(tmp[0], ".")[1])
		}
	}
	return ids, nil
}

func CheckHealth() (string, error) {
	args := []string{"health", "--connect-timeout", "15"}
	return util.ExecCMDWithOutput("ceph", args)
}

func CheckMonStat(num int) (bool, error) {
	args := []string{"mon", "stat", "--connect-timeout", "15"}
	out, err := util.ExecCMDWithOutput("ceph", args)
	if err != nil {
		return false, err
	}
	if strings.Contains(out, "{") && strings.Contains(out, "}") {
		start := strings.IndexAny(out, "{")
		end := strings.IndexAny(out, "}")
		if len(strings.Split(out[start+1:end], ",")) == num {
			return true, nil
		}
	}
	return false, errors.New("Can not check mon status")
}

func GetFSDF(pool string) (string, error) {
	args := []string{"df", "--connect-timeout", "15"}
	out, err := util.ExecCMDWithOutput("ceph", args)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(out), "\n")
	for _, l := range lines {
		if !strings.Contains(l, pool) {
			continue
		}
		return l, nil
	}
	return "", nil
}

func RemoveConf(cli client.Client) error {
	for _, f := range files {
		file := path.Join(root, f)
		_, err := util.ExecCMDWithOutput("rm", []string{"-f", file})
		if err != nil {
			return err
		}
	}
	return nil
}
func SaveConf(cli client.Client) error {
	ctx := context.TODO()
	cm := corev1.ConfigMap{}
	err := cli.Get(ctx, k8stypes.NamespacedName{common.StorageNamespace, global.ConfigMapName}, &cm)
	if err != nil {
		return err
	}
	for _, f := range files {
		date := []byte(cm.Data[f])
		file := path.Join(root, f)
		if err := ioutil.WriteFile(file, date, 0644); err != nil {
			return err
		}
	}
	return nil
}

func CheckConf() bool {
	file := path.Join(root, "ceph.conf")
	_, err := os.Stat(file)
	if err != nil {
		return false
	}
	return true
}
