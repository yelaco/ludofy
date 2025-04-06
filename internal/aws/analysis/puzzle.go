package analysis

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/athena/types"
	"github.com/chess-vn/slchess/internal/domains/dtos"
	"github.com/chess-vn/slchess/internal/domains/entities"
)

func (client *Client) GetPuzzle(
	ctx context.Context,
	id string,
) (
	entities.Puzzle,
	error,
) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE puzzleid = '%s'",
		*client.cfg.PuzzlesTableName,
		id,
	)
	startQueryResp, err := client.athena.StartQueryExecution(
		ctx,
		&athena.StartQueryExecutionInput{
			QueryString: aws.String(query),
			QueryExecutionContext: &types.QueryExecutionContext{
				Database: client.cfg.AthenaDatabaseName,
			},
			ResultConfiguration: &types.ResultConfiguration{
				OutputLocation: client.cfg.PuzzlesResultLocation,
			},
			ResultReuseConfiguration: &types.ResultReuseConfiguration{
				ResultReuseByAgeConfiguration: &types.ResultReuseByAgeConfiguration{
					Enabled:         true,
					MaxAgeInMinutes: aws.Int32(10080),
				},
			},
		})
	if err != nil {
		return entities.Puzzle{}, fmt.Errorf("failed to start query: %w", err)
	}
	if startQueryResp == nil {
		return entities.Puzzle{}, fmt.Errorf("nil start query response: %w", err)
	}

	queryExecutionId := *startQueryResp.QueryExecutionId

	// Wait for query to complete
	for {
		time.Sleep(2 * time.Second)

		statusResp, err := client.athena.GetQueryExecution(ctx, &athena.GetQueryExecutionInput{
			QueryExecutionId: aws.String(queryExecutionId),
		})
		if err != nil {
			return entities.Puzzle{}, fmt.Errorf("failed to get query execution status: %w", err)
		}

		state := statusResp.QueryExecution.Status.State
		if state == types.QueryExecutionStateSucceeded {
			break
		} else if state == types.QueryExecutionStateFailed ||
			state == types.QueryExecutionStateCancelled {
			return entities.Puzzle{}, fmt.Errorf("query failed or was cancelled: %s", *statusResp.QueryExecution.Status.StateChangeReason)
		}
	}

	resultsResp, err := client.athena.GetQueryResults(ctx, &athena.GetQueryResultsInput{
		QueryExecutionId: aws.String(queryExecutionId),
	})
	if err != nil {
		return entities.Puzzle{}, fmt.Errorf("failed to get query results: %w", err)
	}
	if len(resultsResp.ResultSet.Rows) < 2 {
		return entities.Puzzle{}, fmt.Errorf("invalid result")
	}

	columns := []string{}
	for _, col := range resultsResp.ResultSet.Rows[0].Data {
		columns = append(columns, aws.ToString(col.VarCharValue))
	}

	row := resultsResp.ResultSet.Rows[1]
	record := make(map[string]string, len(row.Data))
	for i, colData := range row.Data {
		record[columns[i]] = aws.ToString(colData.VarCharValue)
	}
	puzzle := dtos.PuzzleAthenaQueryResult{
		Id:              record["puzzleid"],
		Fen:             record["fen"],
		Moves:           record["moves"],
		Rating:          record["rating"],
		RatingDeviation: record["ratingdeviation"],
		Popularity:      record["popularity"],
		Nbplays:         record["nbplays"],
		Themes:          record["themes"],
		GameUrl:         record["gameurl"],
		OpeningTags:     record["openingtags"],
	}

	return dtos.PuzzleAthenaQueryToEntity(puzzle), nil
}

func (client *Client) FetchPuzzles(
	ctx context.Context,
	rating float64,
	limit int,
) (
	[]entities.Puzzle,
	error,
) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE rating > %d ORDER BY rating LIMIT %d",
		*client.cfg.PuzzlesTableName,
		int(rating),
		limit,
	)
	startQueryResp, err := client.athena.StartQueryExecution(
		ctx,
		&athena.StartQueryExecutionInput{
			QueryString: aws.String(query),
			QueryExecutionContext: &types.QueryExecutionContext{
				Database: client.cfg.AthenaDatabaseName,
			},
			ResultConfiguration: &types.ResultConfiguration{
				OutputLocation: client.cfg.PuzzlesResultLocation,
			},
		})
	if err != nil {
		return nil, fmt.Errorf("failed to start query: %w", err)
	}
	if startQueryResp == nil {
		return nil, fmt.Errorf("nil start query response: %w", err)
	}

	queryExecutionId := *startQueryResp.QueryExecutionId

	// Wait for query to complete
	for {
		time.Sleep(2 * time.Second)

		statusResp, err := client.athena.GetQueryExecution(ctx, &athena.GetQueryExecutionInput{
			QueryExecutionId: aws.String(queryExecutionId),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get query execution status: %w", err)
		}

		state := statusResp.QueryExecution.Status.State
		if state == types.QueryExecutionStateSucceeded {
			break
		} else if state == types.QueryExecutionStateFailed ||
			state == types.QueryExecutionStateCancelled {
			return nil, fmt.Errorf("query failed or was cancelled: %s", *statusResp.QueryExecution.Status.StateChangeReason)
		}
	}

	resultsResp, err := client.athena.GetQueryResults(ctx, &athena.GetQueryResultsInput{
		QueryExecutionId: aws.String(queryExecutionId),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get query results: %w", err)
	}
	if len(resultsResp.ResultSet.Rows) < 2 {
		return nil, fmt.Errorf("invalid result")
	}

	columns := make([]string, 0, len(resultsResp.ResultSet.Rows[0].Data))
	for _, col := range resultsResp.ResultSet.Rows[0].Data {
		columns = append(columns, aws.ToString(col.VarCharValue))
	}

	record := make(map[string]string, len(columns))
	puzzles := make([]entities.Puzzle, 0, len(resultsResp.ResultSet.Rows))
	for _, row := range resultsResp.ResultSet.Rows[1:] {
		for i, colData := range row.Data {
			record[columns[i]] = aws.ToString(colData.VarCharValue)
		}
		puzzle := dtos.PuzzleAthenaQueryResult{
			Id:              record["puzzleid"],
			Fen:             record["fen"],
			Moves:           record["moves"],
			Rating:          record["rating"],
			RatingDeviation: record["ratingdeviation"],
			Popularity:      record["popularity"],
			Nbplays:         record["nbplays"],
			Themes:          record["themes"],
			GameUrl:         record["gameurl"],
			OpeningTags:     record["openingtags"],
		}
		puzzles = append(puzzles, dtos.PuzzleAthenaQueryToEntity(puzzle))
	}

	return puzzles, nil
}
