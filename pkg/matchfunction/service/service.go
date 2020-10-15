package service

import (
	"github.com/Octops/agones-discover-openmatch/internal/runtime"
	"github.com/Octops/agones-discover-openmatch/pkg/matchfunction/functions"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"open-match.dev/open-match/pkg/matchfunction"
	"open-match.dev/open-match/pkg/pb"
)

type MatchFunctionService struct {
	logger             *logrus.Entry
	queryServiceClient pb.QueryServiceClient
	makeMatchesFunc    functions.MakeMatchesFunc
}

func NewMatchFunctionService(queryServiceClient pb.QueryServiceClient, makeMatchesFunc functions.MakeMatchesFunc) pb.MatchFunctionServer {
	return &MatchFunctionService{
		logger:             runtime.Logger().WithField("source", "match_function"),
		queryServiceClient: queryServiceClient,
		makeMatchesFunc:    makeMatchesFunc,
	}
}

func (s *MatchFunctionService) Run(req *pb.RunRequest, stream pb.MatchFunction_RunServer) error {
	poolTickets, err := matchfunction.QueryPools(stream.Context(), s.queryServiceClient, req.GetProfile().GetPools())
	if err != nil {
		err = errors.Wrap(err, "failed to query pools")
		s.logger.Error(err)
		return err
	}

	proposals, err := s.makeMatchesFunc(req.GetProfile(), poolTickets)
	if err != nil {
		err = errors.Wrap(err, "failed to make matches")
		s.logger.Error(err)
		return err
	}

	for _, proposal := range proposals {
		if err := stream.Send(&pb.RunResponse{Proposal: proposal}); err != nil {
			err := errors.Wrap(err, "failed to stream proposals to Open Match")
			s.logger.Error(err)
			return err
		}
	}

	return nil
}
