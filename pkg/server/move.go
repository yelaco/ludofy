package server

func NewDefaultMove(playerId string) Move {
	return &DefaultMove{
		PlayerId: playerId,
	}
}

func (m *DefaultMove) GetPlayerId() string {
	return m.PlayerId
}
