package nfs

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/gok8s/client"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	mountPath = "/nfs"
)

func StatusControl(cli client.Client, name string) {
	ctrlName := fmt.Sprintf("[%s-nfs-status-controller]", name)
	log.Debugf("%s Start", ctrlName)
	for {
		time.Sleep(60 * time.Second)
		nfs, err := getNfs(cli, name)
		if err != nil {
			if apierrors.IsNotFound(err) == false {
				log.Warnf("%s Get storage failed. Err: %s", ctrlName, err.Error())
				continue
			}
			if err := uMountTmpdir(name); err != nil {
				log.Warnf("%s umount failed. Err: %s", ctrlName, err.Error())
			}
			log.Debugf("%s Stop", ctrlName)
			return
		}
		if nfs.DeletionTimestamp != nil {
			if err := uMountTmpdir(name); err != nil {
				log.Warnf("%s umount failed. Err: %s", ctrlName, err.Error())
			}
			log.Debugf("%s Stop", ctrlName)
			return
		}

		if nfs.Status.Phase == storagev1.Updating || nfs.Status.Phase == storagev1.Creating || nfs.Status.Phase == storagev1.Failed {
			continue
		}
		size, err := getSize(cli, nfs)
		if err != nil {
			log.Warnf("%s Get status failed. Err: %s", ctrlName, err.Error())
			continue
		}
		if err := updateSize(cli, name, size); err != nil {
			log.Warnf("%s Update storage failed. Err: %s", ctrlName, err.Error())
			continue
		}
	}
}

func getSize(cli client.Client, nfs *storagev1.Nfs) (*storagev1.Size, error) {
	targetDir, err := makeMountTargetDir(nfs.Name)
	if err != nil {
		return nil, err
	}
	if err := mountTmpdir(nfs.Spec.Server, nfs.Spec.Path, targetDir); err != nil {
		return nil, err
	}
	return getMountPointSize(targetDir)
}

func makeMountTargetDir(name string) (string, error) {
	dir := filepath.Join(mountPath, name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return dir, os.MkdirAll(dir, os.ModePerm)
	}
	return dir, nil
}

func mountTmpdir(server, path, targetDir string) error {
	nfs := fmt.Sprintf("%s:%s", server, path)
	ok, err := hasMount(nfs, targetDir)
	if err != nil {
		return err
	}
	if !ok {
		if _, err := exec.Command("mount", "-t", "nfs", "-o", "rw,nolock", nfs, targetDir).Output(); err != nil {
			return err
		}
	}
	return nil
}

func uMountTmpdir(name string) error {
	targetDir, err := makeMountTargetDir(name)
	if err != nil {
		return err
	}
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return nil
	}
	out, err := exec.Command("df", targetDir).Output()
	if err != nil {
		return err
	}
	for _, l := range strings.Split(string(out), "\n") {
		if strings.Contains(l, targetDir) {
			if _, err := exec.Command("umount", targetDir).Output(); err != nil {
				return err
			}
		}
	}
	return nil
}

func getMountPointSize(targetDir string) (*storagev1.Size, error) {
	out, err := exec.Command("df", targetDir).Output()
	if err != nil {
		return nil, err
	}
	for _, l := range strings.Split(string(out), "\n") {
		if strings.Contains(l, targetDir) {
			line := strings.Fields(l)
			tsize, err := strconv.ParseInt(line[len(line)-5], 10, 64)
			if err != nil {
				return nil, err
			}
			usize, err := strconv.ParseInt(line[len(line)-4], 10, 64)
			if err != nil {
				return nil, err
			}
			fsize, err := strconv.ParseInt(line[len(line)-3], 10, 64)
			if err != nil {
				return nil, err
			}
			return &storagev1.Size{
				Total: strconv.FormatInt(tsize*1024, 10),
				Used:  strconv.FormatInt(usize*1024, 10),
				Free:  strconv.FormatInt(fsize*1024, 10),
			}, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("can not find mount point %s", targetDir))
}

func hasMount(nfs, targetDir string) (bool, error) {
	out, err := exec.Command("df", targetDir).Output()
	if err != nil {
		return false, err
	}
	for _, l := range strings.Split(string(out), "\n") {
		if strings.Contains(l, nfs) {
			return true, nil
		}
	}
	return false, nil
}
