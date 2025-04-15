package server

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/chess-vn/slchess/pkg/logging"
	"go.uber.org/zap"
)

func sha256Hash(payload []byte) string {
	hash := sha256.Sum256(payload)
	return hex.EncodeToString(hash[:])
}

func signRequestWithSigV4(
	ctx context.Context,
	cfg aws.Config,
	req *http.Request,
) error {
	signer := v4.NewSigner()

	payload, err := io.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}
	req.Body = io.NopCloser(bytes.NewReader(payload)) // Reset body

	// Sign request
	credentials, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		logging.Error("Failed to save game", zap.Error(err))
	}
	err = signer.SignHTTP(
		ctx,
		credentials,
		req,
		sha256Hash(payload),
		"appsync",
		cfg.Region,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to sign request: %w", err)
	}

	return nil
}
