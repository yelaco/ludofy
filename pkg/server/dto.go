package server

type PlayerRecord map[string]string

func (pr PlayerRecord) GetPlayerId() string {
	return pr["PlayerId"]
}

func (pr PlayerRecord) ContainsPlayerId() bool {
	_, ok := pr["PlayerId"]
	return ok
}
