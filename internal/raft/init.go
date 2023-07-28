package raft

import (
	"go-raft-kv/internal/logprovider"
	"go.uber.org/zap"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"time"
)

var log, _ = logprovider.CreateDefaultZapLogger(zap.InfoLevel)

func newNode(address string) *node {
	return &node{
		connect: false,
		address: address,
	}
}

func Init(id int, nodeAddr []string) *Raft {
	ns := make(map[int]*node)
	for k, v := range nodeAddr {
		ns[k] = newNode(v)
	}

	// 创建节点
	return &Raft{
		self:  id,
		nodes: ns,
	}
}

func Start(raft *Raft) {
	raft.state = Follower
	raft.currTerm = 0
	raft.votedFor = -1
	raft.heartbeatC = make(chan bool)
	raft.toLeaderC = make(chan bool)

	go func() {
		r := rand.New(rand.NewSource(time.Now().UnixMilli()))

		for {
			switch raft.state {
			case Follower:
				select {
				case <-raft.heartbeatC:
					log.Info("follower-" + strconv.Itoa(raft.self) + " received heartbeat")
				case <-time.After(time.Duration(r.Intn(5000)+10000) * time.Millisecond):
					log.Info("follower-" + strconv.Itoa(raft.self) + " timeout")
					raft.state = Candidate
				}
			case Candidate:
				log.Info("candidate-" + strconv.Itoa(raft.self) + " running")
				raft.currTerm++
				raft.votedFor = raft.self
				raft.voteCount = 1
				// 申请投票
				go raft.broadcastRequestVote()
				select {
				case <-time.After(time.Duration(r.Intn(5000)+3000) * time.Millisecond):
					log.Info("candidate-" + strconv.Itoa(raft.self) + " timeout")
					raft.state = Follower
					raft.votedFor = -1
				case <-raft.toLeaderC:
					log.Info("candidate-" + strconv.Itoa(raft.self) + " became leader")
					raft.state = Leader
				}
			case Leader:
				raft.broadcastHeartbeat()
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

func (raft *Raft) Rpc(port string) {
	err := rpc.Register(raft)
	if err != nil {
		log.Fatal("rpc register failed", zap.Error(err))
	}
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", port)
	if e != nil {
		log.Fatal("listen error:", zap.Error(err))
	}
	go func() {
		err := http.Serve(l, nil)
		if err != nil {
			log.Fatal("http server error:", zap.Error(err))
		}
	}()
}
