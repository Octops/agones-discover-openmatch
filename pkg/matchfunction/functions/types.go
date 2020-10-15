package functions

import (
	"errors"
	"open-match.dev/open-match/pkg/pb"
)

var (
	ErrMatchProfileIsNil = errors.New("MatchProfile can't be nil")
	ErrPoolTicketsIsNil  = errors.New("PoolTickets can't be nil")
)

type MakeMatchesFunc func(profile *pb.MatchProfile, poolTickets map[string][]*pb.Ticket) ([]*pb.Match, error)
