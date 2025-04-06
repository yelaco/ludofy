package dtos

import (
	"strconv"

	"github.com/chess-vn/slchess/internal/domains/entities"
)

type PuzzleAthenaQueryResult struct {
	Id              string `json:"puzzleid"`
	Fen             string `json:"fen"`
	Moves           string `json:"moves"`
	Rating          string `json:"rating"`
	RatingDeviation string `json:"ratingdeviation"`
	Popularity      string `json:"popularity"`
	Nbplays         string `json:"nbplays"`
	Themes          string `json:"themes"`
	GameUrl         string `json:"gameurl"`
	OpeningTags     string `json:"openingtags"`
}

type PuzzleSolveResponse struct {
	NewRating float64 `json:"newRating"`
}

type PuzzleResponse struct {
	Id              string `json:"puzzleid"`
	Fen             string `json:"fen"`
	Moves           string `json:"moves"`
	Rating          int64  `json:"rating"`
	RatingDeviation int64  `json:"ratingdeviation"`
	Popularity      int64  `json:"popularity"`
	Nbplays         int64  `json:"nbplays"`
	Themes          string `json:"themes"`
	GameUrl         string `json:"gameurl"`
	OpeningTags     string `json:"openingtags"`
}

type PuzzleListResponse struct {
	Items []PuzzleResponse `json:"items"`
}

func PuzzleAthenaQueryToEntity(puzzle PuzzleAthenaQueryResult) entities.Puzzle {
	rating, _ := strconv.ParseInt(puzzle.Rating, 10, 64)
	ratingDeviation, _ := strconv.ParseInt(puzzle.Rating, 10, 64)
	popularity, _ := strconv.ParseInt(puzzle.Rating, 10, 64)
	nbplays, _ := strconv.ParseInt(puzzle.Rating, 10, 64)
	return entities.Puzzle{
		Id:              puzzle.Id,
		Fen:             puzzle.Fen,
		Moves:           puzzle.Moves,
		Rating:          rating,
		RatingDeviation: ratingDeviation,
		Popularity:      popularity,
		Nbplays:         nbplays,
		Themes:          puzzle.Themes,
		GameUrl:         puzzle.GameUrl,
		OpeningTags:     puzzle.OpeningTags,
	}
}

func PuzzleResponseFromEntity(puzzle entities.Puzzle) PuzzleResponse {
	return PuzzleResponse{
		Id:              puzzle.Id,
		Fen:             puzzle.Fen,
		Moves:           puzzle.Moves,
		Rating:          puzzle.Rating,
		RatingDeviation: puzzle.RatingDeviation,
		Popularity:      puzzle.Popularity,
		Nbplays:         puzzle.Nbplays,
		Themes:          puzzle.Themes,
		GameUrl:         puzzle.GameUrl,
		OpeningTags:     puzzle.OpeningTags,
	}
}

func PuzzleListResponseFromEntities(puzzles []entities.Puzzle) PuzzleListResponse {
	puzzleResps := []PuzzleResponse{}
	for _, puzzle := range puzzles {
		puzzleResps = append(puzzleResps, PuzzleResponseFromEntity(puzzle))
	}
	return PuzzleListResponse{
		Items: puzzleResps,
	}
}
