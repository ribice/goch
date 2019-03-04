package main

import (
	"log"

	"github.com/ribice/goch/pkg/config"

	"github.com/ribice/goch/pkg/nats"
	"github.com/ribice/goch/pkg/redis"
	"github.com/ribice/msv"
	authmw "github.com/ribice/msv/middleware/auth"
)

func main() {
	cfg, err := config.Load("./.d")
	checkErr(err)
	nc, err := nats.New(cfg.NATS.ClusterID, cfg.NATS.ClientID, cfg.NATS.URL)
	checkErr(err)
	rc, err := redis.New(cfg.Redis.Address, cfg.Redis.Password, cfg.Redis.Port)
	checkErr(err)
	_, _ = nc, rc

	srv, mux := msv.New(cfg.Server.Port)
	aMW := authmw.New(cfg.Admin.Username, cfg.Admin.Password)
	mux.Use(aMW.WithBasic)

	srv.Start()
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
