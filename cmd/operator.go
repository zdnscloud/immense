package main

import (
	"fmt"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/cement/signal"

	"github.com/zdnscloud/gok8s/client/config"
	"github.com/zdnscloud/immense/pkg/controller"
)

func main() {
	log.InitLogger(log.Debug)

	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatalf(fmt.Sprintf("get config failed:%v\n", err))
	}

	_, err = controller.New(cfg)
	if err != nil {
		log.Fatalf(fmt.Sprintf("create controller failed:%v\n", err))
	}

	signal.WaitForInterrupt(nil)
}
