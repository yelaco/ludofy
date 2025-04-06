package lichess

import "github.com/chess-vn/slchess/internal/domains/entities"

func EvaluationLichessToEntity(eval Evaluation) entities.Evaluation {
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
