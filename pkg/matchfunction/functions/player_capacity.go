package functions

import (
	"context"
	"errors"
	"fmt"
	"github.com/Octops/agones-discover-openmatch/internal/runtime"
	"open-match.dev/open-match/pkg/pb"
	"time"
)

const (
	MATCFUNC_NAME = "player_capacity_matchfunc"
)

var (
	ErrPlayersCapacityInvalid = errors.New("player capacity must be higher than zero")
)

/*
Criteria for Matches
- Number or tickets should not exceed the PlayerCapacity set by the Status.Players.Capacity field from the GS
*/
func MatchByGamePlayersCapacity(playerCapacity int) MakeMatchesFunc {
	return func(profile *pb.MatchProfile, poolTickets map[string][]*pb.Ticket) ([]*pb.Match, error) {
		if err := ValidateMatchFunArguments(playerCapacity, profile, poolTickets); err != nil {
			return nil, err
		}

		//t := []*pb.Ticket{}
		//for _, tickets := range poolTickets {
		//	for _, ticket := range tickets {
		//		t = append(t, ticket)
		//	}
		//}
		//
		//id := fmt.Sprintf("profile-%v-time-%v", profile.GetName(), uuid.New().String())
		//matches := append([]*pb.Match{}, CreateMatchForTickets(id, profile.GetName(), t...))
		//runtime.Logger().Debugf("total matches for profile %s tickets %d: %d", profile.GetName(), len(matches), len(t))
		//return matches, nil

		ctx, cancel := context.WithCancel(context.Background())
		chTickets := make(chan *pb.Ticket)

		go func(pool map[string][]*pb.Ticket) {
			defer func() {
				cancel()
			}()

			for _, tickets := range pool {
				for _, ticket := range tickets {
					t := ticket
					chTickets <- t
				}
			}
		}(poolTickets)

		var tickets []*pb.Ticket
		var matches []*pb.Match
		var match *pb.Match

		for {
			select {
			case t := <-chTickets:
				runtime.Logger().Debugf("creating match for ticket %s", t.Id)
				if tickets == nil || len(match.Tickets) == playerCapacity {
					tickets = []*pb.Ticket{}
					tickets = append(tickets, t)
					id := fmt.Sprintf("profile-%v-%v", profile.GetName(), time.Now().UnixNano())
					matches = append(matches, CreateMatchForTickets(id, profile.GetName(), tickets...))
					match = matches[len(matches)-1]
					break
				}

				if len(match.Tickets) < playerCapacity {
					match.Tickets = append(match.Tickets, t)
					break
				}
			case <-ctx.Done():
				runtime.Logger().Debugf("total matches for profile %s: %d", profile.GetName(), len(matches))
				timeout, _ := context.WithTimeout(context.Background(), time.Second)
				<-timeout.Done()
				return matches, nil
			}
		}
	}
}

func CreateMatchForTickets(matchID, profileName string, tickets ...*pb.Ticket) *pb.Match {
	return &pb.Match{
		MatchId:       matchID,
		MatchProfile:  profileName,
		MatchFunction: MATCFUNC_NAME,
		Tickets:       tickets,
	}
}

func ValidateMatchFunArguments(playerCapacity int, profile *pb.MatchProfile, poolTickets map[string][]*pb.Ticket) error {
	if playerCapacity <= 0 {
		return ErrPlayersCapacityInvalid
	}

	if profile == nil {
		return ErrMatchProfileIsNil
	}

	if poolTickets == nil {
		return ErrPoolTicketsIsNil
	}

	return nil
}
