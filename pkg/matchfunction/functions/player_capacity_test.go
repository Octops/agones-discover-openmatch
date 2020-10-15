package functions

import (
	"github.com/stretchr/testify/require"
	"open-match.dev/open-match/pkg/pb"
	"testing"
)

/*
- PlayerCapacity <= 0 return error PlayerCapacity can't be lower than zero
- Profile == nil return error Profile is nil
- PoolTickets == nil return error PoolTickets is nil
- PoolTickets == 0 return matches == 0
- PoolTickets <= capacity return 1 match with all tickets from the pool
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
		name            string
		playersCapacity int
		profile         *pb.MatchProfile
		poolTickets     map[string][]*pb.Ticket
		wantMatches     int
		wantErr         wantErr
	}{
		{
			name:            "it should return error if PlayerCapacity is lower than zero",
			playersCapacity: -1,
			profile:         nil,
			poolTickets:     nil,
			wantErr: wantErr{
				want: true,
				err:  ErrPlayersCapacityInvalid,
			},
		},
		{
			name:            "it should return error if MatchProfile is nil",
			playersCapacity: 1,
			profile:         nil,
			poolTickets:     nil,
			wantErr: wantErr{
				want: true,
				err:  ErrMatchProfileIsNil,
			},
		},
		{
			name:            "it should return error if PoolTicket is nil",
			playersCapacity: 1,
			profile: &pb.MatchProfile{
				Name: "pool_mode_world",
			},
			poolTickets: nil,
			wantErr: wantErr{
				want: true,
				err:  ErrPoolTicketsIsNil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			matches, err := MatchByGamePlayersCapacity(tc.playersCapacity)(tc.profile, tc.poolTickets)
			if tc.wantErr.want {
				require.Error(t, err)
				require.Equal(t, tc.wantErr.err, err)
			} else {
				require.NoError(t, err)
				require.Len(t, matches, tc.wantMatches)
			}
		})
	}
}
