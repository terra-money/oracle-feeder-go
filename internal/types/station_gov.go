package types

type VoteOption string

const (
	VoteOptionYes VoteOption = "VOTE_OPTION_YES"
	VoteOptionNo  VoteOption = "VOTE_OPTION_NO"
)

type VoteOptionWeight struct {
	Option VoteOption `json:"option"`
	Weight string     `json:"weight"`
}

type StationVote struct {
	Voter   string             `json:"voter"`
	Options []VoteOptionWeight `json:"options"`
}

type StationVotes = []StationVote
