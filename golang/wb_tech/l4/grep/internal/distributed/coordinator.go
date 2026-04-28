package distributed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"mygrep/internal/protocol"
	"mygrep/internal/search"
)

type CoordinatorConfig struct {
	Peers       []string
	Quorum      int
	ShardSize   int
	Parallelism int
	Timeout     time.Duration
}

type shard struct {
	index int
	lines []search.Line
}

type shardVote struct {
	resp protocol.SearchResponse
	err  error
}

type shardResult struct {
	index   int
	matches []search.Line
	err     error
}

func Run(ctx context.Context, lines []search.Line, opts search.Options, cfg CoordinatorConfig) ([]search.Line, error) {
	if len(cfg.Peers) == 0 {
		return search.Filter(lines, opts)
	}

	peers := normalizePeers(cfg.Peers)
	if cfg.Quorum <= 0 {
		cfg.Quorum = len(peers)/2 + 1
	}
	if cfg.Quorum > len(peers) {
		return nil, fmt.Errorf("quorum %d is greater than number of peers %d", cfg.Quorum, len(peers))
	}
	if cfg.ShardSize <= 0 {
		cfg.ShardSize = 128
	}
	if cfg.Parallelism <= 0 {
		cfg.Parallelism = runtime.NumCPU()
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 3 * time.Second
	}

	shards := shardLines(lines, cfg.ShardSize)
	if len(shards) == 0 {
		return nil, nil
	}

	client := &http.Client{Timeout: cfg.Timeout}
	results := make([][]search.Line, len(shards))

	workCh := make(chan shard)
	resultCh := make(chan shardResult, len(shards))

	var workers sync.WaitGroup
	for i := 0; i < cfg.Parallelism; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for shard := range workCh {
				matches, err := queryShard(ctx, client, peers, cfg.Quorum, shard, opts)
				resultCh <- shardResult{
					index:   shard.index,
					matches: matches,
					err:     err,
				}
			}
		}()
	}

	go func() {
		defer close(workCh)
		for _, shard := range shards {
			select {
			case workCh <- shard:
			case <-ctx.Done():
				return
			}
		}
	}()

	var firstErr error
	for range shards {
		select {
		case res := <-resultCh:
			if res.err != nil && firstErr == nil {
				firstErr = res.err
			}
			results[res.index] = res.matches
		case <-ctx.Done():
			firstErr = ctx.Err()
		}
	}

	workers.Wait()

	if firstErr != nil {
		return nil, firstErr
	}

	merged := make([]search.Line, 0)
	for _, shardMatches := range results {
		merged = append(merged, shardMatches...)
	}

	return merged, nil
}

func shardLines(lines []search.Line, shardSize int) []shard {
	if len(lines) == 0 {
		return nil
	}

	shards := make([]shard, 0, (len(lines)+shardSize-1)/shardSize)
	for start, idx := 0, 0; start < len(lines); start, idx = start+shardSize, idx+1 {
		end := start + shardSize
		if end > len(lines) {
			end = len(lines)
		}
		shards = append(shards, shard{
			index: idx,
			lines: append([]search.Line(nil), lines[start:end]...),
		})
	}

	return shards
}

func queryShard(ctx context.Context, client *http.Client, peers []string, quorum int, shard shard, opts search.Options) ([]search.Line, error) {
	request := protocol.SearchRequest{
		ShardIndex: shard.index,
		Options:    opts,
		Lines:      shard.lines,
	}

	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	votes := make(chan shardVote, len(peers))
	for _, peer := range peers {
		go func(peer string) {
			resp, err := callPeer(childCtx, client, peer, request)
			votes <- shardVote{resp: resp, err: err}
		}(peer)
	}

	signatures := make(map[string]int)
	candidates := make(map[string][]search.Line)
	errorsSeen := make([]error, 0)

	for range peers {
		select {
		case vote := <-votes:
			if vote.err != nil {
				errorsSeen = append(errorsSeen, vote.err)
				continue
			}

			signatures[vote.resp.Signature]++
			if _, ok := candidates[vote.resp.Signature]; !ok {
				candidates[vote.resp.Signature] = vote.resp.Matches
			}

			if signatures[vote.resp.Signature] >= quorum {
				cancel()
				return candidates[vote.resp.Signature], nil
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return nil, fmt.Errorf("shard %d did not reach quorum %d: %s", shard.index, quorum, joinErrors(errorsSeen))
}

func callPeer(ctx context.Context, client *http.Client, peer string, request protocol.SearchRequest) (protocol.SearchResponse, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return protocol.SearchResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, peer+"/grep", bytes.NewReader(body))
	if err != nil {
		return protocol.SearchResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return protocol.SearchResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		payload, _ := io.ReadAll(resp.Body)
		return protocol.SearchResponse{}, fmt.Errorf("%s returned %s: %s", peer, resp.Status, strings.TrimSpace(string(payload)))
	}

	var searchResp protocol.SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return protocol.SearchResponse{}, fmt.Errorf("decode response from %s: %w", peer, err)
	}

	return searchResp, nil
}

func normalizePeers(peers []string) []string {
	result := make([]string, 0, len(peers))
	for _, peer := range peers {
		peer = strings.TrimSpace(peer)
		if peer == "" {
			continue
		}
		if !strings.HasPrefix(peer, "http://") && !strings.HasPrefix(peer, "https://") {
			peer = "http://" + peer
		}
		result = append(result, strings.TrimRight(peer, "/"))
	}
	return result
}

func joinErrors(errs []error) string {
	if len(errs) == 0 {
		return "not enough identical responses"
	}

	parts := make([]string, 0, len(errs))
	for _, err := range errs {
		parts = append(parts, err.Error())
	}
	return strings.Join(parts, "; ")
}
