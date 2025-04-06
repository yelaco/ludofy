package storage

import (
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type Client struct {
	dynamodb *dynamodb.Client
	cfg      config
}

type config struct {
	ConnectionsTableName            *string
	UserProfilesTableName           *string
	UserRatingsTableName            *string
	UserMatchesTableName            *string
	MatchmakingTicketsTableName     *string
	ActiveMatchesTableName          *string
	MatchStatesTableName            *string
	MatchRecordsTableName           *string
	MatchResultsTableName           *string
	MessagesTableName               *string
	SpectatorConversationsTableName *string
	UserConversationsTableName      *string
	PuzzleProfilesTableName         *string
	EvaluationsTableName            *string
	FriendshipsTableName            *string
	FriendRequestsTableName         *string
	ApplicationEndpointsTableName   *string
}

func NewClient(dynamoClient *dynamodb.Client) *Client {
	return &Client{
		dynamodb: dynamoClient,
		cfg:      loadConfig(),
	}
}

func loadConfig() config {
	var cfg config
	if v, ok := os.LookupEnv("CONNECTIONS_TABLE_NAME"); ok {
		cfg.ConnectionsTableName = aws.String(v)
	}
	if v, ok := os.LookupEnv("USER_PROFILES_TABLE_NAME"); ok {
		cfg.UserProfilesTableName = aws.String(v)
	}
	if v, ok := os.LookupEnv("USER_RATINGS_TABLE_NAME"); ok {
		cfg.UserRatingsTableName = aws.String(v)
	}
	if v, ok := os.LookupEnv("USER_MATCHES_TABLE_NAME"); ok {
		cfg.UserMatchesTableName = aws.String(v)
	}
	if v, ok := os.LookupEnv("MATCHMAKING_TICKETS_TABLE_NAME"); ok {
		cfg.MatchmakingTicketsTableName = aws.String(v)
	}
	if v, ok := os.LookupEnv("ACTIVE_MATCHES_TABLE_NAME"); ok {
		cfg.ActiveMatchesTableName = aws.String(v)
	}
	if v, ok := os.LookupEnv("MATCH_STATES_TABLE_NAME"); ok {
		cfg.MatchStatesTableName = aws.String(v)
	}
	if v, ok := os.LookupEnv("MATCH_RECORDS_TABLE_NAME"); ok {
		cfg.MatchRecordsTableName = aws.String(v)
	}
	if v, ok := os.LookupEnv("MATCH_RESULTS_TABLE_NAME"); ok {
		cfg.MatchResultsTableName = aws.String(v)
	}
	if v, ok := os.LookupEnv("MESSAGES_TABLE_NAME"); ok {
		cfg.MessagesTableName = aws.String(v)
	}
	if v, ok := os.LookupEnv("SPECTATOR_CONVERSATIONS_TABLE_NAME"); ok {
		cfg.SpectatorConversationsTableName = aws.String(v)
	}
	if v, ok := os.LookupEnv("USER_CONVERSATIONS_TABLE_NAME"); ok {
		cfg.UserConversationsTableName = aws.String(v)
	}
	if v, ok := os.LookupEnv("PUZZLE_PROFILES_TABLE_NAME"); ok {
		cfg.PuzzleProfilesTableName = aws.String(v)
	}
	if v, ok := os.LookupEnv("EVALUATIONS_TABLE_NAME"); ok {
		cfg.EvaluationsTableName = aws.String(v)
	}
	if v, ok := os.LookupEnv("FRIENDSHIPS_TABLE_NAME"); ok {
		cfg.FriendshipsTableName = aws.String(v)
	}
	if v, ok := os.LookupEnv("FRIEND_REQUESTS_TABLE_NAME"); ok {
		cfg.FriendRequestsTableName = aws.String(v)
	}
	if v, ok := os.LookupEnv("APPLICATION_ENDPOINTS_TABLE_NAME"); ok {
		cfg.ApplicationEndpointsTableName = aws.String(v)
	}
	return cfg
}
