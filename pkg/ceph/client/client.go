package client

import (
	"encoding/json"
	"github.com/zdnscloud/immense/pkg/ceph/util"
	"strings"
)

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

func InOsd(id string) error {
	args := []string{"osd", "in", id, "--connect-timeout", "15"}
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

func GetUpAndOutOsdIDs() ([]string, error) {
	ids := make([]string, 0)
	args := []string{"osd", "dump", "--connect-timeout", "15"}
	out, err := util.ExecCMDWithOutput("ceph", args)
	if err != nil {
		return ids, err
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
	args := []string{"health", "--connect-timeout", "15"}
	return util.ExecCMDWithOutput("ceph", args)
}

func GetMon() (MonDump, error) {
	var res MonDump
	args := []string{"mon", "dump", "-f", "json", "--connect-timeout", "15"}
	out, err := util.ExecCMDWithOutput("ceph", args)
	if err != nil {
		return res, err
	}
	json.Unmarshal([]byte(out), &res)
	return res, nil
}

func GetDF() (Df, error) {
	var res Df
	args := []string{"osd", "df", "-f", "json", "--connect-timeout", "15"}
	out, err := util.ExecCMDWithOutput("ceph", args)
	if err != nil {
		return res, err
	}
	json.Unmarshal([]byte(out), &res)
	return res, nil
}

func GetIDToHost(id string) (string, error) {
	args := []string{"osd", "status", "--connect-timeout", "15"}
	out, err := util.ExecCMDWithOutput("ceph", args)
	if err != nil {
		return "", err
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
