package main

import (
	"flag"
	"log"

	"github.com/ribice/msv/middleware/bauth"

	"github.com/ribice/goch/internal/chat"

	"github.com/ribice/goch/internal/agent"

	"github.com/ribice/goch/internal/broker"
	"github.com/ribice/goch/internal/ingest"

	"github.com/ribice/goch/pkg/config"

	"github.com/ribice/goch/pkg/nats"
	"github.com/ribice/goch/pkg/redis"
	"github.com/ribice/msv"
)

func main() {
	cfgPath := flag.String("config", "./conf.yaml", "Path to config file")
	flag.Parse()
	cfg, err := config.Load(*cfgPath)
	checkErr(err)
	mq, err := nats.New(cfg.NATS.ClusterID, cfg.NATS.ClientID, cfg.NATS.URL)
	checkErr(err)
	store, err := redis.New(cfg.Redis.Address, cfg.Redis.Password, cfg.Redis.Port)
	checkErr(err)

	srv, mux := msv.New("goch")
	aMW := bauth.New(cfg.Admin.Username, cfg.Admin.Password, "GOCH")

	agent.NewAPI(mux, broker.New(mq, store, ingest.New(mq, store)), store, cfg)
	chat.New(mux, store, cfg, aMW.MWFunc)

	srv.Start()
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
