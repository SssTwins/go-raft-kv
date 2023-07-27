package raft

import (
	"go.uber.org/zap"
	"net/rpc"
)

type Heartbeat struct {
	// 当前leader term
	Term uint64
	// 当前 leader id
	LeaderId int
}

type HeartbeatReply struct {
	// 当前 follower term
	Term uint64
}

func (raft *Raft) sendHeartbeat(id int, hb Heartbeat, reply *HeartbeatReply) {
	client, err := rpc.DialHTTP("tcp", raft.nodes[id].address)
	if err != nil {
		log.Fatal("dialing: ", zap.Error(err))
		return
	}

	defer func(client *rpc.Client) {
		err := client.Close()
		if err != nil {
			log.Fatal("client close err: ", zap.Error(err))
		}
	}(client)

	callErr := client.Call("Raft.Heartbeat", hb, reply)
	if callErr != nil {
		log.Fatal("dialing: ", zap.Error(callErr))
	}

	if reply.Term > raft.currTerm {
		raft.currTerm = reply.Term
		raft.state = Follower
		raft.votedFor = -1
	}
}

// Heartbeat rpc method
func (raft *Raft) Heartbeat(hb Heartbeat, reply *HeartbeatReply) error {

	// 如果leader节点term小于follower节点，不做处理并返回
	if hb.Term < raft.currTerm {
		reply.Term = raft.currTerm
		return nil
	}

	// 如果leader节点term大于follower节点
	// 说明 follower 过时，重置follower节点term
	if hb.Term > raft.currTerm {
		raft.currTerm = hb.Term
		raft.state = Follower
		raft.votedFor = -1
	}

	// 将当前follower节点term返回给leader
	reply.Term = raft.currTerm

	// 心跳成功，将消息发给 heartbeatC 通道
	raft.heartbeatC <- true

	return nil
}
