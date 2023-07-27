package raft

import (
	"go-raft-kv/internal/logprovider"
	"go.uber.org/zap"
	"math/rand"
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

func start(raft *Raft) {
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
				case <-time.After(time.Duration(r.Intn(500-300)+300) * time.Millisecond):
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
				case <-time.After(time.Duration(r.Intn(5000-300)+300) * time.Millisecond):
					log.Info("candidate-" + strconv.Itoa(raft.self) + " timeout")
					raft.state = Follower
				case <-raft.heartbeatC:
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
