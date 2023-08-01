package v1

import "go.uber.org/zap"

type State = uint8

// Raft节点三种状态：Follower、Candidate、Leader
const (
	Follower State = iota + 1
	PreCandidate
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

	log []LogEntry

	// 被提交日志最大索引
	commitIndex int

	// 已应用到状态机的最大索引
	lastApply int

	// 需要发送给每个节点的下一个索引
	nextIndex []int

	// 已经发送给每个节点的最大索引
	matchIndex []int
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
	for i := range raft.nodes {
		// request 参数
		hb := Heartbeat{
			Term:        raft.currTerm,
			LeaderId:    raft.self,
			CommitIndex: raft.commitIndex,
		}

		prevLogIndex := raft.nextIndex[i] - 1

		// 如果有日志未同步则发送
		if raft.getLastIndex() > prevLogIndex {
			hb.PrevLogIndex = prevLogIndex
			hb.PrevLogTerm = raft.log[prevLogIndex].CurrTerm
			hb.Entries = raft.log[prevLogIndex:]
			log.Info("will send log entries", zap.Any("logEntries", hb.Entries))
		}

		go func(index int, hb Heartbeat) {
			var reply HeartbeatReply
			// 向某一个节点发送 heartbeat
			raft.sendHeartbeat(index, hb, &reply)
		}(i, hb)
	}
}

// 获取日志同步的最后一个index
func (raft *Raft) getLastIndex() int {
	rlen := len(raft.log)
	if rlen == 0 {
		return 0
	}
	return raft.log[rlen-1].Index
}

// 获取日志同步的最后一个term
func (raft *Raft) getLastTerm() uint64 {
	rlen := len(raft.log)
	if rlen == 0 {
		return 0
	}
	return raft.log[rlen-1].CurrTerm
}
