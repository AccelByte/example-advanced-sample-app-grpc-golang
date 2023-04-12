package server

const crewType = "mercenary"

type CrewRules struct {
	AllianceRule AllianceRule `json:"alliance" bson:"allianceRule"`
}

type AllianceRule struct {
	MinNumber       int `json:"min_number" valid:"range(0|2147483647)"`
	MaxNumber       int `json:"max_number" valid:"range(0|2147483647)"`
	PlayerMinNumber int `json:"player_min_number" valid:"range(0|2147483647)"`
	PlayerMaxNumber int `json:"player_max_number" valid:"range(0|2147483647)"`
}
