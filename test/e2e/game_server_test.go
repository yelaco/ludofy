package e2e

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"testing"
	"time"

	"github.com/chess-vn/slchess/internal/domains/dtos"
	"github.com/chess-vn/slchess/pkg/logging"
	"github.com/chess-vn/slchess/pkg/utils"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

type SyncMessages struct {
	Type string `json:"type"`
}

type Message struct {
	Type      string    `json:"type"`
	Data      Data      `json:"data"`
	CreatedAt time.Time `json:"createdAt"`
}

type Data struct {
	Action string `json:"action"`
	Move   string `json:"move"`
}

type GameState struct {
	Type string `json:"type"`
	Game Game   `json:"game"`
}

type Game struct {
	Outcome string   `json:"outcome"`
	Method  string   `json:"method"`
	Fen     string   `json:"fen"`
	Clocks  []string `json:"clocks"`
}

var cfg config

func TestMain(m *testing.M) {
	cfg = newConfig()
	os.Exit(m.Run())
}

func TestGameServer(t *testing.T) {
	matchId, serverIp := testMatchmaking(t)
	time.Sleep(10 * time.Second)

	if os.Getenv("LOCAL") != "" {
		serverIp = "localhost"
	}
	matchUrl := fmt.Sprintf("ws://%s:7202/game/%s", serverIp, matchId)
	player1Header := http.Header{}
	player1Header.Set("Authorization", cfg.User1IdToken)
	player1Conn, _, err := websocket.DefaultDialer.Dial(matchUrl, player1Header)
	require.NoError(t, err)
	defer player1Conn.Close()

	player2Header := http.Header{}
	player2Header.Set("Authorization", cfg.User2IdToken)
	player2Conn, _, err := websocket.DefaultDialer.Dial(matchUrl, player2Header)
	require.NoError(t, err)
	defer player2Conn.Close()

	// Handle OS interrupts for graceful shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	wg := sync.WaitGroup{}
	wg.Add(2)

	player1Msgs, player2Msgs := getTestMessages()
	go func() {
		defer wg.Done()
		var gameState GameState
		require.NoError(t, player1Conn.ReadJSON(&gameState))

		i := 0
		for {
			fen, err := utils.ParseFEN(gameState.Game.Fen)
			if err != nil {
				logging.Fatal(err.Error())
			}
			if i == len(player1Msgs) {
				time.Sleep(2 * time.Second)
				return
			}
			msg := player1Msgs[i]
			if fen.ActiveColor == "w" {
				require.NoError(t, player1Conn.WriteJSON(msg))
				i += 1
			}
			err = player1Conn.ReadJSON(&gameState)
			if errors.Is(err, websocket.ErrCloseSent) {
				player1Conn.WriteControl(websocket.CloseNormalClosure, nil, time.Now().Add(5*time.Second))
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		var gameState GameState
		require.NoError(t, player2Conn.ReadJSON(&gameState))

		i := 0
		for {
			fen, err := utils.ParseFEN(gameState.Game.Fen)
			require.NoError(t, err)
			if i == len(player2Msgs) {
				time.Sleep(2 * time.Second)
				return
			}
			msg := player2Msgs[i]
			if fen.ActiveColor == "b" {
				require.NoError(t, player2Conn.WriteJSON(msg))
				i += 1
			}
			err = player2Conn.ReadJSON(&gameState)
			if errors.Is(err, websocket.ErrCloseSent) {
				player2Conn.WriteControl(websocket.CloseNormalClosure, nil, time.Now().Add(5*time.Second))
				return
			}
		}
	}()

	wg.Wait()
}

func testMatchmaking(t *testing.T) (string, string) {
	client := http.Client{}
	matchmakingReq := getMatchmakingRequest()
	matchmakingReqJson, err := json.Marshal(matchmakingReq)
	require.NoError(t, err)

	apiUrl, err := url.Parse(cfg.ApiUrl)
	require.NoError(t, err)
	matchmakingUrl := apiUrl.JoinPath("matchmaking")

	user2MatchmakingRequest, err := http.NewRequest(http.MethodPost, matchmakingUrl.String(), bytes.NewBuffer(matchmakingReqJson))
	require.NoError(t, err)
	user2MatchmakingRequest.Header.Add("Content-Type", "application/json")
	user2MatchmakingRequest.Header.Add("Authorization", cfg.User2IdToken)
	resp, err := client.Do(user2MatchmakingRequest)
	require.NoError(t, err)
	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		var activeMatchResp dtos.ActiveMatchResponse
		err = json.Unmarshal(body, &activeMatchResp)
		require.NoError(t, err)
		return activeMatchResp.MatchId, activeMatchResp.Server
	}
	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	user1MatchmakingRequest, err := http.NewRequest(http.MethodPost, matchmakingUrl.String(), bytes.NewBuffer(matchmakingReqJson))
	require.NoError(t, err)
	user1MatchmakingRequest.Header.Add("Content-Type", "application/json")
	user1MatchmakingRequest.Header.Add("Authorization", cfg.User1IdToken)
	resp, err = client.Do(user1MatchmakingRequest)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var activeMatchResp dtos.ActiveMatchResponse
	err = json.Unmarshal(body, &activeMatchResp)
	require.NoError(t, err)

	return activeMatchResp.MatchId, activeMatchResp.Server
}

func getMatchmakingRequest() dtos.MatchmakingRequest {
	return dtos.MatchmakingRequest{
		MinRating: 1100,
		MaxRating: 1300,
		GameMode:  "10+0",
	}
}

func getTestMessages() ([]Message, []Message) {
	moves := []string{"e2e4", "e7e5", "f1c4", "b8c6", "d1h5", "g8f6", "h5f7"}
	player1Messages := make([]Message, 0, 4)
	player2Messages := make([]Message, 0, 3)
	for i, move := range moves {
		msg := Message{
			Type: "gameData",
			Data: Data{
				Action: "move",
				Move:   move,
			},
			CreatedAt: time.Now(),
		}
		if i%2 == 0 {
			player1Messages = append(player1Messages, msg)
		} else {
			player2Messages = append(player2Messages, msg)
		}
	}
	return player1Messages, player2Messages
}

func getSyncMessage() SyncMessages {
	return SyncMessages{
		Type: "sync",
	}
}
