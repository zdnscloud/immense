package main

import (
	"encoding/json"
	"fmt"
	"github.com/zdnscloud/immense/pkg/ceph/client"
)

const j = `{"epoch":7,"fsid":"8a46abc7-4526-43e4-a6bf-d88cfeee031b","modified":"2019-07-08 13:49:36.355266","created":"2019-07-08 13:34:22.569989","features":{"persistent":["kraken","luminous","mimic","osdmap-prune"],"optional":[]},"mons":[{"rank":0,"name":"ceph-mon-2","addr":"10.42.3.15:6789/0","public_addr":"10.42.3.15:6789/0"},{"rank":1,"name":"ceph-mon-1","addr":"10.42.4.6:6789/0","public_addr":"10.42.4.6:6789/0"},{"rank":2,"name":"ceph-mon-0","addr":"10.42.5.12:6789/0","public_addr":"10.42.5.12:6789/0"}],"quorum":[0,1,2]}`

func main() {
	var res client.MonDump
	json.Unmarshal([]byte(j), &res)
	fmt.Println(res.Mons)
}
