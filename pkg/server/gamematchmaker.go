// Copyright (c) 2022 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/rand"
	"matchmaking-function-grpc-plugin-server-go/pkg/matchmaker"
	"matchmaking-function-grpc-plugin-server-go/pkg/player"
	"math"
	"sort"
	"sync"
	"time"
)

// New returns a MatchMaker of the MatchLogic interface
func NewGameMatchmaker() MatchLogic {
	return GameMatchMaker{}
}

type queue struct {
	tickets []matchmaker.Ticket
	lock    sync.Mutex
}

func (q *queue) push(ticket matchmaker.Ticket) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.tickets = append(q.tickets, ticket)
}

func (q *queue) pop() *matchmaker.Ticket {
	if len(q.tickets) == 0 {
		return nil
	}
	q.lock.Lock()
	defer q.lock.Unlock()
	ticket := q.tickets[0]
	q.tickets = q.tickets[1:]
	return &ticket
}

func newQueue() *queue {
	return &queue{tickets: make([]matchmaker.Ticket, 0)}
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

	spawnLocation, ok := matchTicket.TicketAttributes["spawnLocation"].(float64)
	if !ok {
		return false, errors.New("spawnLocation must be a non-nil float64 value")
	}
	if spawnLocation == 0.0 {
		return false, errors.New("spawnLocation cannot be nil value for a float")
	}

	logrus.Info("Ticket Validation successful")
	return true, nil
}

// EnrichTicket is responsible for adding logic to the match ticket before match making
func (g GameMatchMaker) EnrichTicket(matchTicket matchmaker.Ticket, ruleSet interface{}) (ticket matchmaker.Ticket, err error) {
	logrus.Info("GAME MATCHMAKER: enrich ticket")
	rand.Seed(uint64(time.Now().UnixNano()))
	var num float64
	enrichMap := map[string]interface{}{}
	if len(matchTicket.TicketAttributes) == 0 {
		logrus.Info("GAME MATCHMAKER: ticket attributes are empty, lets add some!")
		num = float64(rand.Intn(100-0+1) + 0)
		enrichMap["spawnLocation"] = math.Round(num)
		matchTicket.TicketAttributes = enrichMap
		logrus.Infof("EnrichedTicket Attributes: %+v", matchTicket.TicketAttributes)
	} else {
		num = float64(rand.Intn(100-0+1) + 0)
		matchTicket.TicketAttributes["spawnLocation"] = math.Round(num)
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

	go func() {
		var unmatchedTickets []matchmaker.Ticket
		tickets := ticketProvider.GetTickets()
		for ticket := range tickets {
			unmatchedTickets = append(unmatchedTickets, ticket)
			logrus.Errorf("TICKET LENGTH: %d", len(unmatchedTickets))
		}
		go buildGame(unmatchedTickets, results, rules)
	}()

	return results
}

func buildGame(unmatchedTickets []matchmaker.Ticket, results chan matchmaker.Match, gameRules GameRules) {
	defer close(results)
	max := gameRules.AllianceRule.PlayerMaxNumber
	min := gameRules.AllianceRule.PlayerMinNumber
	buckets := map[int]*queue{}
	for _, ticket := range unmatchedTickets {
		bucket, ok := buckets[len(ticket.Players)]
		if !ok {
			bucket = newQueue()
			buckets[len(ticket.Players)] = bucket
		}
		bucket.push(ticket)
	}

	//start outer loop
	for {
		rootTicket := nextTicket(buckets, max)
		if rootTicket == nil {
			return
		}
		remainingPlayerCount := max - len(rootTicket.Players)

		matchedTickets := []matchmaker.Ticket{*rootTicket}

		//start inner loop
		for {
			if remainingPlayerCount == 0 {
				break
			}
			otherTicket := nextTicket(buckets, remainingPlayerCount)
			if otherTicket == nil {
				if remainingPlayerCount >= min {
					break
				}
				return
			}
			matchedTickets = append(matchedTickets, *otherTicket)
			remainingPlayerCount -= len(otherTicket.Players)

		}

		ffaTeam := mapPlayerIDs(matchedTickets)
		match := matchmaker.Match{Tickets: matchedTickets,
			Teams: []matchmaker.Team{{UserIDs: ffaTeam}}}
		results <- match
	}
}

func nextTicket(buckets map[int]*queue, maxPlayerCount int) *matchmaker.Ticket {
	bucketKeys := maps.Keys(buckets)
	sort.Ints(bucketKeys)

	for i := len(bucketKeys) - 1; i >= 0; i-- {
		if bucketKeys[i] > maxPlayerCount {
			continue
		}
		ticket := buckets[bucketKeys[i]].pop()
		if ticket != nil {
			return ticket
		}
	}
	return nil
}

func mapPlayerIDs(tickets []matchmaker.Ticket) []player.ID {
	playerIDs := []player.ID{}
	for _, ticket := range tickets {
		for _, p := range ticket.Players {
			playerIDs = append(playerIDs, p.PlayerID)
		}
	}
	return playerIDs
}
