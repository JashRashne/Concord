package protocol

const (
	MessageTypeMessage             = "MESSAGE"
	MessageTypePing                = "PING"
	MessageTypePong                = "PONG"
	MessageTypeSet                 = "SET"
	MessageTypeGet                 = "GET"
	MessageTypeValue               = "VALUE"
	MessageTypeNotFound            = "NOT_FOUND"
	MessageTypeDelete              = "DELETE"
	MessageTypeOK                  = "OK"
	MessageTypeRequestVote         = "REQUEST_VOTE"
	MessageTypeRequestVoteResponse = "REQUEST_VOTE_RESPONSE"
)

type Message struct {
	Type  string `json:"type"`
	From  string `json:"from"`
	Data  string `json:"data,omitempty"`
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`

	RequestVote *RequestVoteRequest `json:"request_vote,omitempty"`

	RequestVoteResponse *RequestVoteResponse `json:"request_vote_response,omitempty"`
}

type RequestVoteRequest struct {
	Term         uint64 `json:"term"`
	CandidateID  string `json:"candidate_id"`
	LastLogIndex uint64 `json:"last_log_index"`
	LastLogTerm  uint64 `json:"last_log_term"`
}

type RequestVoteResponse struct {
	Term        uint64 `json:"term"`
	VoteGranted bool   `json:"vote_granted"`
}
