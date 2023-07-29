package raft

import (
	"fmt"
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
		r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(raft.self)))
		for {
			switch raft.state {
			case Follower:
				select {
				case <-raft.heartbeatC:
					log.Debug("follower-" + strconv.Itoa(raft.self) + " received heartbeat")
				case <-time.After(time.Duration(r.Intn(500)+1000) * time.Millisecond):
					log.Info("follower-" + strconv.Itoa(raft.self) + " timeout, became candidate")
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
				case <-time.After(time.Duration(r.Intn(150)+150) * time.Millisecond):
					log.Info("candidate-" + strconv.Itoa(raft.self) + " timeout, became follower")
					raft.state = Follower
					raft.votedFor = -1
				case <-raft.toLeaderC:
					log.Info("candidate-" + strconv.Itoa(raft.self) + " became leader")
					raft.state = Leader

					// 初始化日志发送index
					// todo 暂时不考虑leader重新选举的情况
					raft.nextIndex = make([]int, len(raft.nodes))
					raft.matchIndex = make([]int, len(raft.nodes))
					for i := range raft.nodes {
						raft.nextIndex[i] = 1
						raft.matchIndex[i] = 0
					}

					// 模拟客户端发送消息
					go func() {
						i := 0
						for {
							i++
							// leader收到客户端的command，封装成LogEntry并append 到 log
							raft.log = append(raft.log, LogEntry{raft.currTerm, i,
								fmt.Sprintf("user send : %d", i)})
							time.Sleep(3 * time.Second)
						}
					}()
				}
			case Leader:
				raft.broadcastHeartbeat()
				time.Sleep(50 * time.Millisecond)
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
