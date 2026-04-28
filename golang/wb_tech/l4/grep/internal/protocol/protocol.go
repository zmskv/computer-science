package protocol

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"mygrep/internal/search"
)

type SearchRequest struct {
	ShardIndex int            `json:"shardIndex"`
	Options    search.Options `json:"options"`
	Lines      []search.Line  `json:"lines"`
}

type SearchResponse struct {
	ShardIndex int           `json:"shardIndex"`
	Matches    []search.Line `json:"matches"`
	Signature  string        `json:"signature"`
	Error      string        `json:"error,omitempty"`
}

func Signature(lines []search.Line) string {
	payload, err := json.Marshal(lines)
	if err != nil {
		return ""
	}

	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}
