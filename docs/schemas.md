---
id: Database-Design
aliases: []
tags: []
---

### User

```json
{
  "UserId": "player_123",
  "Gmail" "zuka@gmail.com",
  "Password": hash()
  "Username": "zuka",
  "Avatar": "s3://asdfdf",
  "Country": "Vietnam",
  "Elo": 1900

  "SubscriptionStatus": "guest", # guest/premium
  "SubscriptionStartDate: "2025-01-01T00:00:00Z",

}
```

### Active Session

```json
{
  "SessionId": "game_123",
  "Player1": "player_1_uid",
  "Player2": "player_2_uid"
  "Server": "192.168.0.2",
  "StartTime": "2025-01-11T10:00:00Z"
}
```

### Game result

- Update after both player disconnected or game ended (same thing)

```json
{
  "SessionId": "session_123",
  "Players": {
    "Player1": "player_1_uid",
    "Player2": "player_2_uid"
  },
  "TimeControl": "180+2" # Blizt game 180 sec and 2s increment per move
  "StartTime": "2025-01-11T10:00:00Z",
  "EndTime": "2025-01-11T10:15:00Z",
  "Result": "WHITE_CHECKMATE",
  "Pgn": "s3://game-records/game_123"
  "ElapsedTime": [1000, 5000, 20000] # milisecs
}
```

After game ended, push the game replay data to S3

### Session state: -> Use AppSync to sync session state to Spectator

```json
{
  "sessionId": "session_123",
  "players": {
    "player1": {"timeRemaining": 100000, "lastMoveTimestamp": None, status: "connecting", "LastDisconnectTime": None},
    "player2": {"timeRemaining": 100000, "lastMoveTimestamp": None, status: "connecting", "LastDisconnectTime": None},
  },
  gameState: "PGN"
  "increment": 2000,  # 2 seconds increment
  "delay": 0,          # No delay
  "LastUpdated": "2025-01-11T10:15:00Z",
}
```

### Messages

```json
{
  "MessageId" "mss_123"
  "SessionId": "game_123",
  "Sender": "player_1_uid",
  "Content": "Hello World",
  "CreatedAt": "2025-01-11T10:00:00Z"
}
```
