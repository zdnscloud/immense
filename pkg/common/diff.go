package common

import (
	"github.com/zdnscloud/cement/set"
)

func HostsDiff(oldcfg, newcfg []string) ([]string, []string) {
	oldhosts := set.StringSetFromSlice(oldcfg)
	newhosts := set.StringSetFromSlice(newcfg)
	del := oldhosts.Difference(newhosts).ToSlice()
	add := newhosts.Difference(oldhosts).ToSlice()
	return del, add
}
