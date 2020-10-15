package functions

import (
	"errors"
	"fmt"
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

		var matches []*pb.Match
		count := 0
		for {
			insufficientTickets := false
			matchTickets := []*pb.Ticket{}
			for pool, tickets := range poolTickets {
				if len(tickets) < playerCapacity {
					// This pool is completely drained out. Stop creating matches.
					insufficientTickets = true
					break
				}

				// Remove the Tickets from this pool and add to the match proposal.
				matchTickets = append(matchTickets, tickets[0:playerCapacity]...)
				poolTickets[pool] = tickets[playerCapacity:]
			}

			if insufficientTickets {
				break
			}

			matches = append(matches, &pb.Match{
				MatchId:       fmt.Sprintf("profile-%v-time-%v-%v", profile.GetName(), time.Now().Format(time.RFC3339), count),
				MatchProfile:  profile.GetName(),
				MatchFunction: MATCFUNC_NAME,
				Tickets:       matchTickets,
			})

			count++
		}

		return matches, nil
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
