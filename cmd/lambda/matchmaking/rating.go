package main

import (
	"context"
	"fmt"
	"math"

	"github.com/chess-vn/slchess/internal/domains/entities"
)

// Constants
var q = math.Log(10) / 400 // Glicko scaling constant

// g(RD) function
func g(rd float64) float64 {
	return 1 / math.Sqrt(1+3*q*q*rd*rd/(math.Pi*math.Pi))
}

// Expected score function
func expectedScore(r1, r2, rd2 float64) float64 {
	return 1 / (1 + math.Pow(10, -g(rd2)*(r1-r2)/400))
}

// Update rating and deviation
func calculateNewRating(
	userRating entities.UserRating,
	opponentRatings []entities.UserRating,
	results []float64,
) (
	float64,
	float64,
) {
	if len(opponentRatings) != len(results) {
		panic("Mismatch between opponents and results")
	}

	var d2, sum float64
	for i, opp := range opponentRatings {
		E := expectedScore(userRating.Rating, opp.Rating, opp.RD)
		gRD := g(opp.RD)
		d2 += (q * q * gRD * gRD * E * (1 - E))
		sum += gRD * (results[i] - E)
	}

	d2 = 1 / d2
	newRD := math.Sqrt(1 / (1/(userRating.RD*userRating.RD) + 1/d2))
	newRating := userRating.Rating + (q/(1/(userRating.RD*userRating.RD)+1/d2))*sum

	return newRating, newRD
}

func calculateNewRatings(
	ctx context.Context,
	userRating,
	opponentRating entities.UserRating,
) (
	[]float64,
	[]float64,
	error,
) {
	if deploymentStage == "dev" {
		return []float64{userRating.Rating, userRating.Rating, userRating.Rating},
			[]float64{userRating.RD, userRating.RD, userRating.RD},
			nil
	}

	matchResults, _, err := storageClient.FetchMatchResults(ctx, userRating.UserId, nil, 5)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch match results: %w", err)
	}

	opponentRatings := make([]entities.UserRating, 0, len(matchResults))
	results := make([]float64, len(matchResults)+1)
	for i, matchResult := range matchResults {
		opponentRatings = append(opponentRatings, entities.UserRating{
			UserId: matchResult.OpponentId,
			Rating: matchResult.OpponentRating,
			RD:     matchResult.OpponentRD,
		})

		results[i] = matchResult.Result
	}
	opponentRatings = append(opponentRatings, opponentRating)

	newRatings := make([]float64, 3)
	newRDs := make([]float64, 3)

	possibleResults := []float64{1.0, 0.5, 0.0}
	for i, result := range possibleResults {
		results[len(matchResults)] = result
		newRatings[i], newRDs[i] = calculateNewRating(
			userRating,
			opponentRatings,
			results,
		)
	}

	return newRatings, newRDs, nil
}
