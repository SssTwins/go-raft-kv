package main

import (
	"go-raft-kv/internal/logprovider"
	"go.uber.org/zap"
	"strconv"
)

func main() {
	log, _ := logprovider.CreateDefaultZapLogger(zap.InfoLevel)
	defer log.Error("Hello End!")
	log.Info("Hello World!")
	for i := 0; i < 4; i++ {
		log.Info(strconv.Itoa(i))
	}
}
