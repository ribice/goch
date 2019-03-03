package main

import (
	"log"
	"os"
	"strconv"

	"github.com/ribice/goch/pkg/nats"
	"github.com/ribice/goch/pkg/redis"
	"github.com/ribice/msv"
	authmw "github.com/ribice/msv/middleware/auth"
)

func main() {
	nc, err := nats.New(mustGetEnv("NATS_CLUSTER_ID"), mustGetEnv("NATS_CLIENT_ID"), mustGetEnv("NATS_URL"))

	redisPort, err := strconv.Atoi(mustGetEnv("REDIS_PORT"))
	checkErr(err)
	rc, err := redis.New(mustGetEnv("REDIS_ADDR"), os.Getenv("REDIS_PASS"), redisPort)
	checkErr(err)
	_, _ = nc, rc

	srv, mux := msv.New(8080)
	aMW := authmw.New(mustGetEnv("ADMIN_USER"), mustGetEnv("ADMIN_PASS"))
	mux.Use(aMW.WithBasic)

	srv.Start()
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func mustGetEnv(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		log.Fatalf("env variable %s required but not found", key)
	}
	return v

}
