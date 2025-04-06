package dtos

import (
	"github.com/chess-vn/slchess/internal/domains/entities"
)

type PvResponse struct {
	Cp    int    `json:"cp"`
	Moves string `json:"moves"`
}

type EvaluationResponse struct {
	Fen    string       `json:"fen"`
	Depth  int          `json:"depth"`
	Knodes int          `json:"knodes"`
	Pvs    []PvResponse `json:"pvs"`
}

type EvaluationRequest struct {
	ConnectionId string `json:"connectionId"`
	Fen          string `json:"fen"`
}

type EvaluationWorkResponse struct {
	ConnectionId  string `json:"connectionId"`
	Fen           string `json:"fen"`
	ReceiptHandle string `json:"receiptHandle"`
}

type EvaluationSubmission struct {
	ConnectionId  string           `json:"connectionId"`
	ReceiptHandle string           `json:"receiptHandle"`
	Evaluation    EvaluationResult `json:"evaluation"`
}

type PvResult struct {
	Cp    int    `json:"cp"`
	Moves string `json:"moves"`
}

type EvaluationResult struct {
	Fen    string     `json:"fen"`
	Depth  int        `json:"depth"`
	Knodes int        `json:"knodes"`
	Pvs    []PvResult `json:"pvs"`
}

func EvaluationResultToEntity(eval EvaluationResult) entities.Evaluation {
	v := entities.Evaluation{
		Fen:    eval.Fen,
		Depth:  eval.Depth,
		Knodes: eval.Knodes,
		Pvs:    make([]entities.Pv, 0, len(eval.Pvs)),
	}
	for _, pv := range eval.Pvs {
		v.Pvs = append(v.Pvs, entities.Pv{
			Cp:    pv.Cp,
			Moves: pv.Moves,
		})
	}
	return v
}

func EvaluationWorkFromRequest(req EvaluationRequest) entities.EvaluationWork {
	return entities.EvaluationWork{
		ConnectionId: req.ConnectionId,
		Fen:          req.Fen,
	}
}

func EvaluationWorkResponseFromEntity(work entities.EvaluationWork) EvaluationWorkResponse {
	return EvaluationWorkResponse{
		ConnectionId:  work.ConnectionId,
		Fen:           work.Fen,
		ReceiptHandle: work.ReceiptHandle,
	}
}

func EvaluationResponseFromEntity(eval entities.Evaluation) EvaluationResponse {
	v := EvaluationResponse{
		Fen:    eval.Fen,
		Depth:  eval.Depth,
		Knodes: eval.Knodes,
		Pvs:    make([]PvResponse, 0, len(eval.Pvs)),
	}
	for _, pv := range eval.Pvs {
		v.Pvs = append(v.Pvs, PvResponse{
			Cp:    pv.Cp,
			Moves: pv.Moves,
		})
	}
	return v
}
