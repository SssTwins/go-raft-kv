package test

import (
	"flag"
	"go-raft-kv/internal/raft/v1"
	"strings"
	"testing"
)

func TestNode2(t *testing.T) {
	port := flag.String("port", ":9092", "rpc listen port")
	cluster := flag.String("cluster", "127.0.0.1:9091,127.0.0.1:9093", "comma sep")
	id := flag.Int("id", 2, "node ID")

	// 参数解析
	flag.Parse()
	clusters := strings.Split(*cluster, ",")
	raftNode := v1.Init(*id, clusters)

	// 监听rpc
	raftNode.Rpc(*port)
	// 开启 raft
	v1.Start(raftNode)

	select {}
}
