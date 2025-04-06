package entities

type Puzzle struct {
	Id              string
	Fen             string
	Moves           string
	Rating          int64
	RatingDeviation int64
	Popularity      int64
	Nbplays         int64
	Themes          string
	GameUrl         string
	OpeningTags     string
}
