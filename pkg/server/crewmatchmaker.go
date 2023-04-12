// Copyright (c) 2022 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elliotchance/pie/v2"
	"github.com/sirupsen/logrus"
	"matchmaking-function-grpc-plugin-server-go/pkg/matchmaker"
	"matchmaking-function-grpc-plugin-server-go/pkg/player"
)

const (
	crewKey   = "crewType"
	crewValue = "mercenary"
)

// New returns a MatchMaker of the MatchLogic interface
func NewCrewMatchmaker() MatchLogic {
	return CrewMatchMaker{}
}

// ValidateTicket for Crew is to ensure that we don't have more than the Maximum Player Number rule that is set in the Matchpool
func (c CrewMatchMaker) ValidateTicket(matchTicket matchmaker.Ticket, matchRules interface{}) (bool, error) {
	logrus.Info("CREW MATCHMAKER: validate ticket")
	rules, ok := matchRules.(CrewRules)
	if !ok {
		return false, errors.New("invalid rules type for crew rules")
	}
	if len(matchTicket.Players) > rules.AllianceRule.PlayerMaxNumber {
		return false, errors.New(fmt.Sprintf("max number of players exceeded, "+
			"player count %d, max %d", len(matchTicket.Players),
			rules.AllianceRule.PlayerMaxNumber))

	}
	return true, nil
}

// EnrichTicket for Crew is responsible for adding logic to the match ticket before match making.
func (c CrewMatchMaker) EnrichTicket(matchTicket matchmaker.Ticket, ruleSet interface{}) (ticket matchmaker.Ticket, err error) {
	logrus.Info("CREW MATCHMAKER: enrich ticket")
	return matchTicket, nil
}

// GetStatCodes returns the string slice of the stat codes in matchrules
func (c CrewMatchMaker) GetStatCodes(matchRules interface{}) []string {
	logrus.Infof("CREW MATCHMAKER: stat codes: %s", []string{})
	return []string{}
}

// RulesFromJSON returns the ruleset from the Game rules
func (c CrewMatchMaker) RulesFromJSON(jsonRules string) (interface{}, error) {
	var ruleSet CrewRules
	err := json.Unmarshal([]byte(jsonRules), &ruleSet)
	if err != nil {
		return nil, err
	}
	return ruleSet, nil
}

// MakeMatches iterates over all the match tickets and matches them based on the buildMatch function
func (c CrewMatchMaker) MakeMatches(ticketProvider TicketProvider, matchRules interface{}) <-chan matchmaker.Match {
	logrus.Info("CREW MATCHMAKER: make matches")
	results := make(chan matchmaker.Match)
	rules, ok := matchRules.(CrewRules)
	if !ok {
		logrus.Error("invalid rules type for crew rules")

		return results
	}
	ctx := context.Background()
	go func() {
		defer close(results)
		var unmatchedTickets []matchmaker.Ticket
		nextTicket := ticketProvider.GetTickets()
		for {
			select {
			case ticket, ok := <-nextTicket:
				if !ok {
					logrus.Info("CREW MATCHMAKER: there are no tickets to create a match with")
					return
				}
				logrus.Infof("CREW MATCHMAKER: got a ticket: %s", ticket.TicketID)
				unmatchedTickets = buildCrew(ticket, unmatchedTickets, rules, results)
			case <-ctx.Done():
				logrus.Info("CREW MATCHMAKER: CTX Done triggered")
				return
			}
		}
	}()
	return results
}

// buildCrew is responsible for building a crew of Players based on the MatchPool Rules of the MaxNumber of players for a Crew (aka Team, Party, etc).
// This is driven from the slice of match tickets and then feeds the results to the match channel.
// When a match is found, a map is added to the Match of "crewType":"mercenary"
func buildCrew(ticket matchmaker.Ticket, unmatchedTickets []matchmaker.Ticket, crewRules CrewRules, results chan matchmaker.Match) []matchmaker.Ticket {
	logrus.Info("CREW MATCHMAKER: seeing if we have enough tickets to match")
	unmatchedTickets = append(unmatchedTickets, ticket)
	enrichMap := map[string]interface{}{
		crewKey: crewValue,
	}
	if len(unmatchedTickets) == crewRules.AllianceRule.MaxNumber {
		logrus.Info("CREW MATCHMAKER: I have enough tickets to match!")
		players := append(unmatchedTickets[0].Players, unmatchedTickets[1].Players...)
		playerIDs := pie.Map(players, player.ToID)
		match := matchmaker.Match{
			RegionPreference: []string{"any"},
			Tickets:          make([]matchmaker.Ticket, crewRules.AllianceRule.MaxNumber),
			Teams: []matchmaker.Team{
				{UserIDs: playerIDs},
			},
			MatchAttributes: enrichMap,
		}
		logrus.Infof("Crew Makeup: %s", match.Teams)
		copy(match.Tickets, unmatchedTickets)
		logrus.Info("CREW MATCHMAKER: sending to results channel")
		results <- match
		logrus.Info("CREW MATCHMAKER: resetting unmatched tickets")
		unmatchedTickets = nil
	}
	logrus.Info("CREW MATCHMAKER: not enough tickets to build a match")
	return unmatchedTickets
}
