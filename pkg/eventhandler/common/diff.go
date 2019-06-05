package common

import (
	"github.com/zdnscloud/cement/set"
	storagev1 "github.com/zdnscloud/immense/pkg/apis/zcloud/v1"
	"reflect"
)

func tomap(cfgs *storagev1.Cluster) map[string][]string {
	res := make(map[string][]string)
	for _, v := range cfgs.Spec.Hosts {
		res[v.NodeName] = v.BlockDevices
	}
	return res
}

func Diff(oldcfg, newcfg *storagev1.Cluster) (map[string][]string, map[string][]string, map[string][]string, map[string][]string) {
	oldmap := tomap(oldcfg)
	newmap := tomap(newcfg)
	addcfg := make(map[string][]string)
	delcfg := make(map[string][]string)
	hosts := make([]string, 0)
	for k, v := range oldmap {
		n, ok := newmap[k]
		if !ok {
			delcfg[k] = v
			continue
		}
		if reflect.DeepEqual(v, n) {
			continue
		}
		hosts = append(hosts, k)
	}
	for k, v := range newmap {
		_, ok := oldmap[k]
		if !ok {
			addcfg[k] = v
			continue
		}
	}
	changetodel := make(map[string][]string)
	changetoadd := make(map[string][]string)
	for _, host := range hosts {
		olddevs := set.StringSetFromSlice(oldmap[host])
		newdevs := set.StringSetFromSlice(newmap[host])
		olddiff := olddevs.Difference(newdevs).ToSlice()
		newdiff := newdevs.Difference(olddevs).ToSlice()
		if len(olddiff) != 0 {
			changetodel[host] = olddiff
		}
		if len(newdiff) != 0 {
			changetoadd[host] = newdiff
		}
	}
	return delcfg, addcfg, changetodel, changetoadd
}