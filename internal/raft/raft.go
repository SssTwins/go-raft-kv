package raft

type State = uint8

// Raft节点三种状态：Follower、Candidate、Leader
const (
	Follower State = iota + 1
	Candidate
	Leader
)

type node struct {
	connect bool
	address string
}

type Raft struct {

	// 当前节点id
	self int

	// 除当前节点外其他节点信息
	nodes map[int]*node

	// 当前节点状态
	state State

	// 当前任期
	currTerm uint64

	// 当前任期投票给了谁
	// 未投票设置为-1
	votedFor int

	// 当前任期获取的投票数量
	voteCount int

	// heartbeat channel
	heartbeatC chan bool

	// to leader channel
	toLeaderC chan bool
}

// 广播请求各个节点投票
func (raft *Raft) broadcastRequestVote() {

	var vote = Vote{
		Term:        raft.currTerm,
		CandidateId: raft.self,
	}

	for i := range raft.nodes {
		go func(i int) {
			var reply VoteReply
			// 发送申请到某个节点
			raft.sendRequestVote(i, vote, &reply)
		}(i)
	}
}

func (raft *Raft) broadcastHeartbeat() {
	// 遍历所有节点
	for id := range raft.nodes {
		// request 参数
		hb := Heartbeat{
			Term:     raft.currTerm,
			LeaderId: raft.self,
		}

		go func(id int, hb Heartbeat) {
			var reply HeartbeatReply
			// 向某一个节点发送 heartbeat
			raft.sendHeartbeat(id, hb, &reply)
		}(id, hb)
	}
}
