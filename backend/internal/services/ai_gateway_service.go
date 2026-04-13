package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"wealthscope-backend/internal/models"
)

type AIGatewayService struct {
	BaseURL string
	Client  *http.Client
}

func NewAIGatewayService(baseURL string) *AIGatewayService {
	return &AIGatewayService{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (s *AIGatewayService) Recommend(
	ctx context.Context,
	req models.AIRecommendRequest,
) ([]models.AIRecommendResponse, error) {
	var out []models.AIRecommendResponse
	if err := s.postJSON(ctx, "/recommend", req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *AIGatewayService) postJSON(ctx context.Context, path string, body any, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		s.BaseURL+path,
		bytes.NewReader(payload),
	)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.Client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("ai service returned status %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}
