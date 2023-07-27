package test

import (
	"go-raft-kv/internal/logprovider"
	"go.uber.org/zap"
	"strconv"
	"testing"
)

func TestNode3(t *testing.T) {
	log, _ := logprovider.CreateDefaultZapLogger(zap.InfoLevel)
	defer log.Error("Hello End!")
	log.Info("Hello World!")
	for i := 0; i < 4; i++ {
		log.Info(strconv.Itoa(i))
	}
}
