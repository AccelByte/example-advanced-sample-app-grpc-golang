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

// New returns a MatchMaker of the MatchLogic interface
func NewGameMatchmaker() MatchLogic {
	return GameMatchMaker{}
}

// ValidateTicket returns a bool if the match ticket is valid
func (g GameMatchMaker) ValidateTicket(matchTicket matchmaker.Ticket, matchRules interface{}) (bool, error) {
	logrus.Info("GAME MATCHMAKER: validate ticket")
	rules, ok := matchRules.(GameRules)
	if !ok {
		return false, errors.New("invalid rules type for game rules")
	}

	if len(matchTicket.Players) > rules.AllianceRule.PlayerMaxNumber {
		return false, errors.New(fmt.Sprintf("too many players on the ticket, max is %d", rules.AllianceRule.PlayerMaxNumber))

	}

	logrus.Info("Ticket Validation successful")
	return true, nil
}

// EnrichTicket is responsible for adding logic to the match ticket before match making
func (g GameMatchMaker) EnrichTicket(matchTicket matchmaker.Ticket, ruleSet interface{}) (ticket matchmaker.Ticket, err error) {
	logrus.Info("GAME MATCHMAKER: enrich ticket")
	if len(matchTicket.TicketAttributes) == 0 {
		logrus.Info("GAME MATCHMAKER: ticket attributes are empty, lets add some!")
		enrichMap := map[string]interface{}{
			"enrichedNumber": float64(20),
		}
		matchTicket.TicketAttributes = enrichMap
		logrus.Infof("EnrichedTicket Attributes: %+v", matchTicket.TicketAttributes)
	} else {
		matchTicket.TicketAttributes["gameMMR"] = float64(20)
		logrus.Infof("EnrichedTicket Attributes: %+v", matchTicket.TicketAttributes)
	}
	return matchTicket, nil
}

// GetStatCodes returns the string slice of the stat codes in matchrules
func (g GameMatchMaker) GetStatCodes(matchRules interface{}) []string {
	logrus.Infof("GAME MATCHMAKER: stat codes: %s", []string{})
	return []string{}
}

// RulesFromJSON returns the ruleset from the Game rules
func (g GameMatchMaker) RulesFromJSON(jsonRules string) (interface{}, error) {
	var ruleSet GameRules
	err := json.Unmarshal([]byte(jsonRules), &ruleSet)
	if err != nil {
		return nil, err
	}
	return ruleSet, nil
}

// MakeMatches iterates over all the match tickets and matches them based on the buildMatch function
func (g GameMatchMaker) MakeMatches(ticketProvider TicketProvider, matchRules interface{}) <-chan matchmaker.Match {
	logrus.Info("GAME MATCHMAKER: make matches")
	results := make(chan matchmaker.Match)
	rules, ok := matchRules.(GameRules)
	if !ok {
		logrus.Error("invalid rules type for game rules")
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
					logrus.Info("GAME MATCHMAKER: there are no tickets to create a match with")
					return
				}
				logrus.Infof("GAME MATCHMAKER: got a ticket: %s", ticket.TicketID)
				unmatchedTickets = buildGame(ticket, unmatchedTickets, rules, results)
			case <-ctx.Done():
				logrus.Info("GAME MATCHMAKER: CTX Done triggered")
				return
			}
		}
	}()
	return results
}

// buildGame is responsible for building matches from the slice of match tickets and feeding them to the match channel.
// Here we are taking the Players from the first 2 unmatched tickets and having them be the Teams matched in the Match.
// This will build us a 2v2 Match (which is what the Session Template is set for, as well)
func buildGame(ticket matchmaker.Ticket, unmatchedTickets []matchmaker.Ticket, gameRules GameRules, results chan matchmaker.Match) []matchmaker.Ticket {
	logrus.Info("GAME MATCHMAKER: seeing if we have enough tickets to match")
	unmatchedTickets = append(unmatchedTickets, ticket)
	if len(unmatchedTickets) == gameRules.AllianceRule.MaxNumber {
		logrus.Info("GAME MATCHMAKER: I have enough tickets to match!")
		teamOneIDs := pie.Map(unmatchedTickets[0].Players, player.ToID)
		logrus.Infof("TeamOne: %s", teamOneIDs)
		teamTwoIDs := pie.Map(unmatchedTickets[1].Players, player.ToID)
		logrus.Infof("TeamTwo: %s", teamTwoIDs)
		match := matchmaker.Match{
			Tickets: make([]matchmaker.Ticket, gameRules.AllianceRule.MaxNumber),
			Teams: []matchmaker.Team{
				{UserIDs: teamOneIDs},
				{UserIDs: teamTwoIDs},
			},
		}
		copy(match.Tickets, unmatchedTickets)
		logrus.Infof("Game Makeup: %s", match.Teams)
		logrus.Info("GAME MATCHMAKER: sending to results channel")
		results <- match
		logrus.Info("GAME MATCHMAKER: resetting unmatched tickets")
		unmatchedTickets = nil
	}
	logrus.Info("GAME MATCHMAKER: not enough tickets to build a match")
	return unmatchedTickets
}
