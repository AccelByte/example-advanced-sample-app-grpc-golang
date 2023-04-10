// Copyright (c) 2022 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package server

import (
	"context"
	"encoding/json"
	"github.com/elliotchance/pie/v2"
	"github.com/sirupsen/logrus"
	"matchmaking-function-grpc-plugin-server-go/pkg/matchmaker"
	"matchmaking-function-grpc-plugin-server-go/pkg/player"
)

// New returns a MatchMaker of the MatchLogic interface
func NewCrewMatchmaker() MatchLogic {
	return CrewMatchMaker{}
}

// ValidateTicket returns a bool if the match ticket is valid
func (c CrewMatchMaker) ValidateTicket(matchTicket matchmaker.Ticket, matchRules interface{}) (bool, error) {
	logrus.Info("CREW MATCHMAKER: validate ticket")
	logrus.Info("Ticket Validation successful")
	return true, nil
}

// EnrichTicket is responsible for adding logic to the match ticket before match making
func (c CrewMatchMaker) EnrichTicket(matchTicket matchmaker.Ticket, ruleSet interface{}) (ticket matchmaker.Ticket, err error) {
	logrus.Info("CREW MATCHMAKER: enrich ticket")

	if len(matchTicket.TicketAttributes) == 0 {
		logrus.Info("CREW MATCHMAKER: ticket attributes are empty, lets add some!")
		enrichMap := map[string]interface{}{
			"enrichedNumber": float64(20),
		}
		matchTicket.TicketAttributes = enrichMap
		logrus.Infof("EnrichedTicket Attributes: %+v", matchTicket.TicketAttributes)
	} else {
		matchTicket.TicketAttributes["enrichedNumber"] = float64(20)
		logrus.Infof("EnrichedTicket Attributes: %+v", matchTicket.TicketAttributes)
	}
	return matchTicket, nil
}

// GetStatCodes returns the string slice of the stat codes in matchrules
func (c CrewMatchMaker) GetStatCodes(matchRules interface{}) []string {
	logrus.Infof("CREW MATCHMAKER: stat codes: %s", []string{})
	return []string{}
}

// RulesFromJSON returns the ruleset from the Game rules
func (c CrewMatchMaker) RulesFromJSON(jsonRules string) (interface{}, error) {
	var ruleSet GameRules
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
				unmatchedTickets = buildCrew(ticket, unmatchedTickets, results)
			case <-ctx.Done():
				logrus.Info("CREW MATCHMAKER: CTX Done triggered")
				return
			}
		}
	}()
	return results
}

// buildMatch is responsible for building matches from the slice of match tickets and feeding them to the match channel
func buildCrew(ticket matchmaker.Ticket, unmatchedTickets []matchmaker.Ticket, results chan matchmaker.Match) []matchmaker.Ticket {
	logrus.Info("CREW MATCHMAKER: seeing if we have enough tickets to match")
	unmatchedTickets = append(unmatchedTickets, ticket)
	if len(unmatchedTickets) == 2 {
		logrus.Info("CREW MATCHMAKER: I have enough tickets to match!")
		players := append(unmatchedTickets[0].Players, unmatchedTickets[1].Players...)
		playerIDs := pie.Map(players, player.ToID)
		match := matchmaker.Match{
			RegionPreference: []string{"any"},
			Tickets:          make([]matchmaker.Ticket, 2),
			Teams: []matchmaker.Team{
				{UserIDs: playerIDs},
			},
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
