package raft

import (
	"go.uber.org/zap"
	"net/rpc"
	"strconv"
)

type Vote struct {
	Term        uint64
	CandidateId int
}

type VoteReply struct {
	Term uint64
	// 回应为true则获得投票
	Reply bool
}

func (raft *Raft) sendRequestVote(id int, vote Vote, reply *VoteReply) {
	client, err := rpc.DialHTTP("tcp", raft.nodes[id].address)
	if err != nil {
		log.Error("dialing: ", zap.Error(err))
		return
	}
	defer func(client *rpc.Client) {
		err := client.Close()
		if err != nil {
			log.Error("client close err: ", zap.Error(err))
			return
		}
	}(client)
	callErr := client.Call("Raft.RequestVote", vote, reply)
	if callErr != nil {
		log.Error("dialing: ", zap.Error(callErr))
		return
	}

	// 如果candidate节点term小于follower节点
	// 当前candidate节点无效
	// candidate节点转变为follower节点
	if reply.Term > raft.currTerm {
		raft.currTerm = reply.Term
		raft.state = Follower
		raft.votedFor = -1
		return
	}

	if reply.Reply {
		raft.voteCount++
		log.Info("当前任期: " + strconv.FormatUint(raft.currTerm, 10) + ", 节点id: " + strconv.Itoa(raft.self) + ", 票数: " + strconv.Itoa(raft.voteCount))
		if raft.voteCount > len(raft.nodes)/2+1 {
			raft.toLeaderC <- true
		}
	}
}

// RequestVote Follower处理投票请求的方法
// 因Candidate会先投票给自己，所以此处对节点状态做判断
func (raft *Raft) RequestVote(vote Vote, reply *VoteReply) error {

	// 如果Candidate节点term小于follower
	// 说明Candidate节点过时，拒绝投票
	if vote.Term < raft.currTerm {
		reply.Term = raft.currTerm
		reply.Reply = false
		return nil
	}

	// 未投票的话则给请求节点投票
	if raft.votedFor == -1 {
		raft.currTerm = vote.Term
		raft.votedFor = vote.CandidateId
		reply.Term = raft.currTerm
		reply.Reply = true
		return nil
	}

	reply.Term = raft.currTerm
	reply.Reply = false
	return nil
}
