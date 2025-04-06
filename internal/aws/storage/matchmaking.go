package storage

import (
	"context"

	"github.com/chess-vn/slchess/internal/domains/entities"
)

func (client *Client) CheckForActiveMatch(
	ctx context.Context,
	userId string,
) (
	entities.ActiveMatch,
	error,
) {
	userMatch, err := client.GetUserMatch(ctx, userId)
	if err != nil {
		return entities.ActiveMatch{}, err
	}
	activeMatch, err := client.GetActiveMatch(ctx, userMatch.MatchId)
	if err != nil {
		return entities.ActiveMatch{}, err
	}
	return activeMatch, nil
}
