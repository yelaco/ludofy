package stofinet

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/chess-vn/slchess/internal/domains/dtos"
	"github.com/chess-vn/slchess/pkg/logging"
	"go.uber.org/zap"
)

type Client struct {
	http *http.Client
	cfg  Config
}

func NewClient() *Client {
	cfg, err := LoadConfig()
	if err != nil {
		logging.Fatal("couldn't load config", zap.Error(err))
	}
	return &Client{
		http: new(http.Client),
		cfg:  cfg,
	}
}

func (client *Client) Start(ctx context.Context) error {
	backoffTime := 2 * time.Second
	var (
		work      dtos.EvaluationWorkResponse
		stop      bool
		available bool = true
	)
	for !stop {
		select {
		case <-ctx.Done():
			stop = true
		default:
		}

		if available {
			// If no work is in progress, acquire new work
			w, err := client.AcquireEvaluationWork(ctx)
			if err != nil {
				if errors.Is(err, ErrEvaluationWorkNotFound) {
					time.Sleep(backoffTime)
					continue
				}
				return fmt.Errorf("failed to acquire evaluation work: %w", err)
			}
			work = w
			available = false
		} else {
			eval, err := client.Evaluate(work.Fen, 20)
			if err != nil {
				return fmt.Errorf("failed to evaluate: %w", err)
			}

			sub := dtos.EvaluationSubmission{
				ConnectionId:  work.ConnectionId,
				ReceiptHandle: work.ReceiptHandle,
				Evaluation:    EvaluationResultFromStofinet(eval),
			}

			newWork, err := client.SubmitEvaluation(ctx, sub, stop)
			if err != nil {
				if errors.Is(err, EOF) ||
					errors.Is(err, ErrEvaluationWorkNotFound) {
					available = true
					continue
				}
				return fmt.Errorf("failed to submit evaluation: %w", err)
			}
			work = newWork
		}
	}
	logging.Info("stopped")
	return nil
}

func (client *Client) AcquireEvaluationWork(ctx context.Context) (dtos.EvaluationWorkResponse, error) {
	logging.Info("checking for new evaluation work")
	u := client.cfg.BaseUrl.JoinPath("acquire")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return dtos.EvaluationWorkResponse{}, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := client.http.Do(req)
	if err != nil {
		return dtos.EvaluationWorkResponse{}, fmt.Errorf("failed to send request: %w", err)
	}
	logging.Info("work acquire responded", zap.String("status", resp.Status))

	switch resp.StatusCode {
	case http.StatusAccepted:
		var eval dtos.EvaluationWorkResponse
		if err := json.NewDecoder(resp.Body).Decode(&eval); err != nil {
			return dtos.EvaluationWorkResponse{}, fmt.Errorf("failed to decode body: %w", err)
		}
		return eval, nil
	case http.StatusNoContent:
		return dtos.EvaluationWorkResponse{}, ErrEvaluationWorkNotFound
	default:
		return dtos.EvaluationWorkResponse{}, ErrUnknownStatusCode
	}
}

func (client *Client) SubmitEvaluation(
	ctx context.Context,
	sub dtos.EvaluationSubmission,
	stop bool,
) (
	dtos.EvaluationWorkResponse,
	error,
) {
	// subJson, err:= json.Marshal(sub)
	u := client.cfg.BaseUrl.JoinPath("evaluation")

	bodyJson, err := json.Marshal(sub)
	if err != nil {
		return dtos.EvaluationWorkResponse{}, fmt.Errorf(" failed to marshal body: %w", err)
	}
	body := bytes.NewReader(bodyJson)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), body)
	if err != nil {
		return dtos.EvaluationWorkResponse{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	if stop {
		params := url.Values{}
		params.Add("stop", "true")
		u.RawQuery = params.Encode()
	}

	resp, err := client.http.Do(req)
	if err != nil {
		return dtos.EvaluationWorkResponse{}, fmt.Errorf("failed to send request: %w", err)
	}
	logging.Info(
		"evaluation submitted",
		zap.Int("status_code", resp.StatusCode),
	)

	switch resp.StatusCode {
	case http.StatusOK:
		return dtos.EvaluationWorkResponse{}, EOF
	case http.StatusAccepted:
		var eval dtos.EvaluationWorkResponse
		if err := json.NewDecoder(resp.Body).Decode(&eval); err != nil {
			return dtos.EvaluationWorkResponse{}, fmt.Errorf("failed to decode body: %w", err)
		}
		return eval, nil
	case http.StatusNoContent:
		return dtos.EvaluationWorkResponse{}, ErrEvaluationWorkNotFound
	default:
		return dtos.EvaluationWorkResponse{}, ErrUnknownStatusCode
	}
}

func (client *Client) Evaluate(fen string, depth int) (Evaluation, error) {
	logging.Info("evaluating", zap.String("fen", fen))

	// set stockfish options
	opts := defaultOptions()
	opts.threads = client.cfg.NumThreads
	opts.hash = client.cfg.HashSize

	pvLines, err := runStockfish(client.cfg.StockfishPath, fen, depth, opts)
	if err != nil {
		return Evaluation{}, fmt.Errorf("failed to run stockfish: %w", err)
	}
	eval := parsePvsLines(pvLines)
	eval.Fen = fen

	logging.Info("done evaluating")
	return eval, nil
}
