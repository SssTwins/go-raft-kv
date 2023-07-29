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

	// 新日志之前的索引
	PrevLogIndex int

	// PrevLogIndex 的任期号
	PrevLogTerm uint64

	// 准备存储的日志条目（表示心跳时为空）
	Entries []LogEntry

	// 被提交日志最大索引
	CommitIndex int
}

type HeartbeatReply struct {

	// 当前 follower term
	Term uint64

	// 心跳请求是否成功
	Reply bool

	// 如果 follower index < leader index, 需要通知leader下次开始发送索引的位置
	NextIndex int
}

func (raft *Raft) sendHeartbeat(index int, hb Heartbeat, reply *HeartbeatReply) {
	client, err := rpc.DialHTTP("tcp", raft.nodes[index].address)
	if err != nil {
		log.Error("dialing: ", zap.Error(err))
		return
	}

	defer func(client *rpc.Client) {
		err := client.Close()
		if err != nil {
			log.Error("client close err: ", zap.Error(err))
		}
	}(client)

	callErr := client.Call("Raft.Heartbeat", hb, reply)
	if callErr != nil {
		log.Error("dialing: ", zap.Error(callErr))
	}

	// 心跳请求返回成功才执行
	if reply.Reply {
		if reply.NextIndex > 0 {
			raft.nextIndex[index] = reply.NextIndex
			raft.matchIndex[index] = raft.nextIndex[index] - 1

			// TODO
			// 如果大于半数节点同步成功
			// 1. 更新 leader 节点的 commitIndex
			// 2. 返回给客户端
			// 3. 应用状态就机
			// 4. 通知 Followers Entry 已提交
		}
	} else {

		if reply.Term > raft.currTerm {
			raft.currTerm = reply.Term
			raft.state = Follower
			raft.votedFor = -1
			return
		}

		// TODO
		// 失败的请求需要处理失联过的节点日志

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

	if len(hb.Entries) == 0 {
		reply.Reply = true
		reply.Term = raft.currTerm
		return nil
	}

	// 如果携带日志同步信息
	// leader维护的最大日志index > follower 的最大index
	// 说明follower有过失联，需要重新同步日志
	if hb.PrevLogIndex > raft.getLastIndex() {
		reply.Reply = false
		reply.Term = raft.currTerm
		reply.NextIndex = raft.getLastIndex() + 1
		return nil
	}

	raft.log = append(raft.log, hb.Entries...)
	// 更新已提交的日志index
	raft.commitIndex = raft.getLastIndex()

	reply.Reply = true
	reply.Term = raft.currTerm
	reply.NextIndex = raft.getLastIndex() + 1

	return nil
}
