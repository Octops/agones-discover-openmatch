package frontend

import (
	"context"
	"github.com/Octops/agones-discover-openmatch/internal/runtime"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"open-match.dev/open-match/pkg/pb"
)

type FrontEndService struct {
	logger *logrus.Entry
	conn   *grpc.ClientConn
	client pb.FrontendServiceClient
}

func NewFrontEndService(conn *grpc.ClientConn) (*FrontEndService, error) {
	logger := runtime.Logger()

	fe := pb.NewFrontendServiceClient(conn)
	return &FrontEndService{
		logger: logger,
		conn:   conn,
		client: fe,
	}, nil
}

func (fe *FrontEndService) CreateTicket(ctx context.Context, ticket *pb.CreateTicketRequest, opts ...grpc.CallOption) (*pb.Ticket, error) {
	return fe.client.CreateTicket(ctx, ticket, opts...)
}
