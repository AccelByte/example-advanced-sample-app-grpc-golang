// Copyright (c) 2022 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package server

type GameRules struct {
	AllianceRule AllianceRule `json:"alliance" bson:"allianceRule"`
	CrewType     string       `json:"crewType"`
}
