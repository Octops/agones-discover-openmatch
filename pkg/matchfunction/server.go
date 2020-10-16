package matchfunction

import (
	"context"
	"fmt"
	"github.com/Octops/agones-discover-openmatch/internal/runtime"
	"github.com/Octops/agones-discover-openmatch/pkg/config"
	"github.com/Octops/agones-discover-openmatch/pkg/matchfunction/functions"
	"github.com/Octops/agones-discover-openmatch/pkg/matchfunction/service"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"open-match.dev/open-match/pkg/pb"
)

type MatchFunction interface {
	Run(*pb.RunRequest, pb.MatchFunction_RunServer) error
}

type Server struct {
	logger             *logrus.Entry
	conn               *grpc.ClientConn
	grpcServer         *grpc.Server
	queryServiceClient pb.QueryServiceClient
}

func NewServer() (*Server, error) {
	logger := runtime.Logger().WithField("source", "server")
	return &Server{
		logger:     logger,
		grpcServer: grpc.NewServer(),
	}, nil
}

func (s *Server) DialQueryService(addr string) error {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return errors.Wrapf(err, "error dialing QueryService on %s", addr)
	}

	s.conn = conn
	s.queryServiceClient = pb.NewQueryServiceClient(conn)

	return nil
}

func (s *Server) RegisterMatchFunction(factory func(client pb.QueryServiceClient, makeMatchesFunc functions.MakeMatchesFunc) pb.MatchFunctionServer, makeMatchesFunc functions.MakeMatchesFunc) {
	matchFunctionService := factory(s.queryServiceClient, makeMatchesFunc)
	pb.RegisterMatchFunctionServer(s.grpcServer, matchFunctionService)
}

func (s *Server) Serve(ctx context.Context, port int32) error {
	defer s.Finalizer()

	if err := s.DialQueryService(config.OpenMatch().QueryService); err != nil {
		return errors.Wrap(err, "failed to dial OpenMatch Query Service")
	}

	// TODO: PlayerCapacity 10 is a random number but must match with the GS Status.Players.Capacity
	s.RegisterMatchFunction(service.NewMatchFunctionService, functions.MatchByGamePlayersCapacity(10))

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return errors.Wrapf(err, "TCP net listener initialization failed for port %d", port)
	}

	defer ln.Close()

	ctxServer, cancel := context.WithCancel(ctx)
	defer cancel()

	s.logger.Infof("TCP net listener initialized for port %d", port)
	go func() {
		if err := s.grpcServer.Serve(ln); err != nil {
			s.logger.Fatal(errors.Wrap(err, "gRPC serve failed"))
			cancel()
		}
	}()

	<-ctxServer.Done()
	return nil
}

func (s *Server) Finalizer() {
	s.logger.Info("stopping match function server")
	if s.conn != nil {
		s.conn.Close()
	}
	s.grpcServer.Stop()
}
