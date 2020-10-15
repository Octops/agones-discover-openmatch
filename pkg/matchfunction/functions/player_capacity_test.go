package functions

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"open-match.dev/open-match/pkg/pb"
	"testing"
)

/*
- PlayerCapacity <= 0 return error PlayerCapacity can't be lower than zero
- Profile == nil return error Profile is nil
- PoolTickets == nil return error PoolTickets is nil
- PoolTickets == 0 return matches == 0
- PoolTickets Tickets <= capacity return 1 match with all tickets from the pool
- PoolTickets > capacity
	- tickets % capacity == 0 return matches count == tickets / capacity
	- tickets % capacity != 0 return matches count == (tickets / capacity) + 1
*/
func TestMatchByGamePlayersCapacity(t *testing.T) {
	type wantErr struct {
		want bool
		err  error
	}

	testCases := []struct {
		name             string
		capacity         int
		profile          *pb.MatchProfile
		poolTickets      map[string][]*pb.Ticket
		wantMatches      int
		wantTotalTickets int
		wantErr          wantErr
	}{
		{
			name:        "it should return error if PlayerCapacity is lower than zero",
			capacity:    -1,
			profile:     nil,
			poolTickets: nil,
			wantErr: wantErr{
				want: true,
				err:  ErrPlayersCapacityInvalid,
			},
		},
		{
			name:        "it should return error if MatchProfile is nil",
			capacity:    1,
			profile:     nil,
			poolTickets: nil,
			wantErr: wantErr{
				want: true,
				err:  ErrMatchProfileIsNil,
			},
		},
		{
			name:     "it should return error if PoolTicket is nil",
			capacity: 1,
			profile: &pb.MatchProfile{
				Name: "pool_mode_world",
			},
			poolTickets: nil,
			wantErr: wantErr{
				want: true,
				err:  ErrPoolTicketsIsNil,
			},
		},
		{
			name:     "it should return zero matches if PoolTicket is empty",
			capacity: 1,
			profile: &pb.MatchProfile{
				Name: "pool_mode_world",
			},
			poolTickets: map[string][]*pb.Ticket{},
			wantErr: wantErr{
				want: false,
				err:  nil,
			},
			wantMatches: 0,
		},
		{
			name:     "it should return 1 match with 1 ticket if PlayerCapacity is higher",
			capacity: 2,
			profile: &pb.MatchProfile{
				Name: "pool_mode_world",
			},
			poolTickets: map[string][]*pb.Ticket{
				"pool_1": {
					{
						Id: uuid.New().String(),
					},
				},
			},
			wantErr: wantErr{
				want: false,
				err:  nil,
			},
			wantMatches:      1,
			wantTotalTickets: 1,
		},
		{
			name:     "it should return 1 match with 2 tickets if PlayerCapacity is higher",
			capacity: 2,
			profile: &pb.MatchProfile{
				Name: "pool_mode_world",
			},
			poolTickets: map[string][]*pb.Ticket{
				"pool_1": {
					{
						Id: uuid.New().String(),
					},
				},
				"pool_2": {
					{
						Id: uuid.New().String(),
					},
				},
			},
			wantErr: wantErr{
				want: false,
				err:  nil,
			},
			wantMatches:      1,
			wantTotalTickets: 2,
		},
		{
			name:     "it should return 2 matches with 1 ticket if PlayerCapacity is lower",
			capacity: 1,
			profile: &pb.MatchProfile{
				Name: "pool_mode_world",
			},
			poolTickets: map[string][]*pb.Ticket{
				"pool_1": {
					{
						Id: uuid.New().String(),
					},
				},
				"pool_2": {
					{
						Id: uuid.New().String(),
					},
				},
			},
			wantErr: wantErr{
				want: false,
				err:  nil,
			},
			wantMatches:      2,
			wantTotalTickets: 2,
		},
		{
			name:     "it should return 2 matches with 3 tickets in total if PlayerCapacity is lower",
			capacity: 2,
			profile: &pb.MatchProfile{
				Name: "pool_mode_world",
			},
			poolTickets: map[string][]*pb.Ticket{
				"pool_1": {
					{
						Id: uuid.New().String(),
					},
				},
				"pool_2": {
					{
						Id: uuid.New().String(),
					},
				},
				"pool_3": {
					{
						Id: uuid.New().String(),
					},
				},
			},
			wantErr: wantErr{
				want: false,
				err:  nil,
			},
			wantMatches:      2,
			wantTotalTickets: 3,
		},
		{
			name:     "it should return 2 matches with 4 tickets in total if PlayerCapacity is equal",
			capacity: 2,
			profile: &pb.MatchProfile{
				Name: "pool_mode_world",
			},
			poolTickets: map[string][]*pb.Ticket{
				"pool_1": {
					{
						Id: uuid.New().String(),
					},
				},
				"pool_2": {
					{
						Id: uuid.New().String(),
					},
				},
				"pool_3": {
					{
						Id: uuid.New().String(),
					},
				},
				"pool_4": {
					{
						Id: uuid.New().String(),
					},
				},
			},
			wantErr: wantErr{
				want: false,
				err:  nil,
			},
			wantMatches:      2,
			wantTotalTickets: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			matches, err := MatchByGamePlayersCapacity(tc.capacity)(tc.profile, tc.poolTickets)
			if tc.wantErr.want {
				require.Error(t, err)
				require.Equal(t, tc.wantErr.err, err)
			} else {
				tickets := []*pb.Ticket{}
				for _, match := range matches {
					tickets = append(tickets, match.Tickets...)
				}

				require.NoError(t, err)
				require.Len(t, matches, tc.wantMatches, "Number of Matches is wrong")
				require.Len(t, tickets, tc.wantTotalTickets, "Number of Tickets is wrong")
			}
		})
	}
}
