package lichess

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/chess-vn/slchess/internal/domains/entities"
)

type Pv struct {
	Cp    int    `json:"cp"`
	Moves string `json:"moves"`
}

type Evaluation struct {
	Fen    string `json:"fen"`
	Depth  int    `json:"depth"`
	Knodes int    `json:"knodes"`
	Pvs    []Pv   `json:"pvs"`
}

var ErrEvaluationNotFound = fmt.Errorf("evaluation not found")

func (client *Client) CloudEvaluate(fen string) (entities.Evaluation, error) {
	apiUrl, err := url.Parse(client.ApiUrl)
	if err != nil {
		return entities.Evaluation{}, fmt.Errorf("failed to parse api url: %w", err)
	}
	cloudEvalUrl := apiUrl.JoinPath("cloud-eval")

	params := url.Values{}
	params.Add("fen", fen)
	cloudEvalUrl.RawQuery = params.Encode()

	req, err := http.NewRequest(
		http.MethodGet,
		cloudEvalUrl.String(),
		nil,
	)
	if err != nil {
		return entities.Evaluation{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.http.Do(req)
	if err != nil {
		return entities.Evaluation{}, fmt.Errorf("failed to send request: %w", err)
	}
	switch resp.StatusCode {
	case http.StatusNotFound:
		return entities.Evaluation{}, ErrEvaluationNotFound
	case http.StatusOK:
		var eval Evaluation
		err := json.NewDecoder(resp.Body).Decode(&eval)
		if err != nil {
			return entities.Evaluation{}, fmt.Errorf("failed to decode body: %w", err)
		}
		return EvaluationLichessToEntity(eval), nil
	default:
		return entities.Evaluation{}, fmt.Errorf("unknown response status code")
	}
}
