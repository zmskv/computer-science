package distributed

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"

	"mygrep/internal/protocol"
	"mygrep/internal/search"
)

type job struct {
	req      protocol.SearchRequest
	resultCh chan result
}

type result struct {
	resp protocol.SearchResponse
	err  error
}

type Server struct {
	jobs chan job
}

func NewServer(workers int) *Server {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}

	s := &Server{
		jobs: make(chan job),
	}

	for i := 0; i < workers; i++ {
		go s.worker()
	}

	return s
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/grep", s.handleGrep)
	return mux
}

func (s *Server) worker() {
	for job := range s.jobs {
		matches, err := search.Filter(job.req.Lines, job.req.Options)
		if err != nil {
			job.resultCh <- result{err: err}
			continue
		}

		job.resultCh <- result{
			resp: protocol.SearchResponse{
				ShardIndex: job.req.ShardIndex,
				Matches:    matches,
				Signature:  protocol.Signature(matches),
			},
		}
	}
}

func (s *Server) handleGrep(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	var req protocol.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("decode request: %v", err), http.StatusBadRequest)
		return
	}

	resultCh := make(chan result, 1)

	select {
	case s.jobs <- job{req: req, resultCh: resultCh}:
	case <-r.Context().Done():
		http.Error(w, "request cancelled", http.StatusRequestTimeout)
		return
	}

	select {
	case res := <-resultCh:
		if res.err != nil {
			http.Error(w, res.err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(res.resp); err != nil {
			http.Error(w, fmt.Sprintf("encode response: %v", err), http.StatusInternalServerError)
		}
	case <-r.Context().Done():
		http.Error(w, "request cancelled", http.StatusRequestTimeout)
	}
}

func Serve(ctx context.Context, listenAddr string, workers int) error {
	server := NewServer(workers)

	httpServer := &http.Server{
		Addr:    listenAddr,
		Handler: server.Handler(),
	}

	go func() {
		<-ctx.Done()
		_ = httpServer.Shutdown(context.Background())
	}()

	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
