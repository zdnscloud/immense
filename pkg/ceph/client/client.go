package client

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zdnscloud/immense/pkg/ceph/global"
	"github.com/zdnscloud/immense/pkg/ceph/util"
)

const (
	cephCMD = "/usr/bin/ceph"
)

var cephTimeOut = []string{"--connect-timeout", "15"}

func ReweigtOsd(id string) error {
	args := []string{"osd", "crush", "reweight", id, "0"}
	args = append(args, cephTimeOut...)
	out, err := util.ExecCMDWithOutput(cephCMD, args)
	if err != nil {
		return fmt.Errorf("filed osd cursh reweight,cmd out: %s, err: %v", out, err)
	}
	return nil
}

func OutOsd(id string) error {
	args := []string{"osd", "out", id}
	args = append(args, cephTimeOut...)
	out, err := util.ExecCMDWithOutput(cephCMD, args)
	if err != nil {
		return fmt.Errorf("filed osd out,cmd out: %s, err: %v", out, err)
	}
	return nil
}

func InOsd(id string) error {
	args := []string{"osd", "in", id}
	args = append(args, cephTimeOut...)
	out, err := util.ExecCMDWithOutput(cephCMD, args)
	if err != nil {
		return fmt.Errorf("filed osd in,cmd out: %s, err: %v", out, err)
	}
	return nil
}

func RemoveCrush(id string) error {
	args := []string{"osd", "crush", "remove", id}
	args = append(args, cephTimeOut...)
	out, err := util.ExecCMDWithOutput(cephCMD, args)
	if err != nil {
		return fmt.Errorf("filed osd crush remove,cmd out: %s, err: %v", out, err)
	}
	return nil
}

func RmOsd(id string) error {
	args := []string{"osd", "rm", id}
	args = append(args, cephTimeOut...)
	out, err := util.ExecCMDWithOutput(cephCMD, args)
	if err != nil {
		return fmt.Errorf("filed osd rm,cmd out: %s, err: %v", out, err)
	}
	return nil
}

func RmOsdAuth(id string) error {
	args := []string{"auth", "del", id}
	args = append(args, cephTimeOut...)
	out, err := util.ExecCMDWithOutput(cephCMD, args)
	if err != nil {
		return fmt.Errorf("filed auth del,cmd out: %s, err: %v", out, err)
	}
	return nil
}

func GetDownOsdIDs(stat string) ([]string, error) {
	ids := make([]string, 0)
	args := []string{"osd", "dump"}
	args = append(args, cephTimeOut...)
	out, err := util.ExecCMDWithOutput(cephCMD, args)
	if err != nil {
		return ids, fmt.Errorf("filed osd dump,cmd out: %s, err: %v", out, err)
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

func GetUpAndOutOsdIDs() ([]string, error) {
	ids := make([]string, 0)
	args := []string{"osd", "dump"}
	args = append(args, cephTimeOut...)
	out, err := util.ExecCMDWithOutput(cephCMD, args)
	if err != nil {
		return ids, fmt.Errorf("filed osd dump,cmd out: %s, err: %v", out, err)
	}
	lines := strings.Split(string(out), "\n")
	for _, l := range lines {
		tmp := strings.Fields(l)
		if strings.HasPrefix(tmp[0], "osd") && tmp[1] == "up" && tmp[2] == "out" {
			ids = append(ids, strings.Split(tmp[0], ".")[1])
		}
	}
	return ids, nil
}

func CheckHealth() (string, error) {
	args := []string{"health"}
	args = append(args, cephTimeOut...)
	return util.ExecCMDWithOutput(cephCMD, args)
}

func GetMon() (MonDump, error) {
	var res MonDump
	args := []string{"mon", "dump", "-f", "json"}
	args = append(args, cephTimeOut...)
	out, err := util.ExecCMDWithOutput(cephCMD, args)
	if err != nil {
		return res, fmt.Errorf("filed mon dump,cmd out: %s, err: %v", out, err)
	}
	json.Unmarshal([]byte(out), &res)
	return res, nil
}

func GetDF() (Df, error) {
	var res Df
	args := []string{"osd", "df", "-f", "json"}
	args = append(args, cephTimeOut...)
	out, err := util.ExecCMDWithOutput(cephCMD, args)
	if err != nil {
		return res, fmt.Errorf("filed osd df,cmd out: %s, err: %v", out, err)
	}
	json.Unmarshal([]byte(out), &res)
	return res, nil
}

func GetIDToHost(id string) (string, error) {
	args := []string{"osd", "status"}
	args = append(args, cephTimeOut...)
	out, err := util.ExecCMDWithOutput(cephCMD, args)
	if err != nil {
		return "", fmt.Errorf("filed osd status,cmd out: %s, err: %v", out, err)
	}
	lines := strings.Split(string(out), "\n")
	for _, l := range lines {
		if !strings.Contains(l, "ceph-osd-") {
			continue
		}
		tmp := strings.Fields(l)
		if tmp[1] == id {
			return tmp[3], nil
		}
	}
	return "", nil
}

func GetHostToIDs(host string) ([]string, error) {
	var ids []string
	args := []string{"osd", "status"}
	args = append(args, cephTimeOut...)
	out, err := util.ExecCMDWithOutput(cephCMD, args)
	if err != nil {
		return ids, fmt.Errorf("filed osd status,cmd out: %s, err: %v", out, err)
	}
	lines := strings.Split(string(out), "\n")
	for _, l := range lines {
		if !strings.Contains(l, "ceph-osd-") {
			continue
		}
		tmp := strings.Fields(l)
		if strings.Contains(tmp[3], host) {
			ids = append(ids, tmp[1])
		}
	}
	return ids, nil
}

func GetCurrentSizeOrPgnum(flag string) (string, error) {
	var num string
	args := []string{"osd", "pool", "get", global.CephFsDate, flag}
	args = append(args, cephTimeOut...)
	out, err := util.ExecCMDWithOutput(cephCMD, args)
	if err != nil {
		return num, fmt.Errorf("filed get pool %s size or pg_num, %v", global.CephFsDate, err)
	}
	lines := strings.Split(string(out), "\n")
	for _, l := range lines {
		if strings.HasPrefix(l, flag) {
			return strings.Fields(l)[1], nil
		}
	}
	return num, fmt.Errorf("can not get pool %s size or pg_num", global.CephFsDate)
}

func UpdateSizeOrPgnum(pool, flag string, num string) error {
	args := []string{"osd", "pool", "set", pool, flag, num}
	args = append(args, cephTimeOut...)
	out, err := util.ExecCMDWithOutput(cephCMD, args)
	if err != nil {
		return fmt.Errorf("filed set pool size,cmd out: %s, err: %v", out, err)
	}
	return nil
}

func EnableDashboard() error {
	args := []string{"mgr", "module", "enable", "dashboard"}
	args = append(args, cephTimeOut...)
	if out, err := util.ExecCMDWithOutput(cephCMD, args); err != nil {
		return fmt.Errorf("filed enable dashboard ,cmd out: %s, err: %v", out, err)
	}
	args = []string{"dashboard", "ac-user-show"}
	args = append(args, cephTimeOut...)
	out, err := util.ExecCMDWithOutput(cephCMD, args)
	if err != nil {
		return fmt.Errorf("filed show user,cmd out: %s, err: %v", out, err)
	}
	if !strings.Contains(out, "admin") {
		args = []string{"dashboard", "ac-user-create", "admin", "cephfs", "administrator"}
		args = append(args, cephTimeOut...)
		if out, err := util.ExecCMDWithOutput(cephCMD, args); err != nil {
			return fmt.Errorf("filed create user,cmd out: %s, err: %v", out, err)
		}
	}
	args = []string{"dashboard", "create-self-signed-cert"}
	args = append(args, cephTimeOut...)
	if out, err := util.ExecCMDWithOutput(cephCMD, args); err != nil {
		return fmt.Errorf("filed create cert,cmd out: %s, err: %v", out, err)
	}
	return nil
}
