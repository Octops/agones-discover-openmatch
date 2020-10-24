package openmatch

import (
	"context"
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"open-match.dev/open-match/pkg/pb"
	"testing"
)

func TestCleanUpAssignmentsWithoutConnection(t *testing.T) {
	type args struct {
		group []*pb.AssignmentGroup
	}

	testCases := []struct {
		name string
		args args
		want []*pb.AssignmentGroup
	}{
		{
			name: "it should not clean up assignment with connection",
			args: args{
				group: []*pb.AssignmentGroup{
					{
						Assignment: &pb.Assignment{
							Connection: "66.211.39.62:7000",
						},
					},
				},
			},
			want: []*pb.AssignmentGroup{
				{
					Assignment: &pb.Assignment{
						Connection: "66.211.39.62:7000",
					},
				},
			},
		},
		{
			name: "it should clean up 1 assignment with no connection",
			args: args{
				group: []*pb.AssignmentGroup{
					{
						Assignment: &pb.Assignment{},
					},
				},
			},
			want: nil,
		},
		{
			name: "it should clean up 1 assignment and keep 1 with connection",
			args: args{
				group: []*pb.AssignmentGroup{
					{
						Assignment: &pb.Assignment{},
					},
					{
						Assignment: &pb.Assignment{
							Connection: "66.211.39.62:7000",
						},
					},
				},
			},
			want: []*pb.AssignmentGroup{
				{
					Assignment: &pb.Assignment{
						Connection: "66.211.39.62:7000",
					},
				},
			},
		},
		{
			name: "it should clean up 1 assignment and keep 2 with connection",
			args: args{
				group: []*pb.AssignmentGroup{
					{
						Assignment: &pb.Assignment{},
					},
					{
						Assignment: &pb.Assignment{
							Connection: "66.211.39.62:7000",
						},
					},
					{
						Assignment: &pb.Assignment{
							Connection: "45.211.39.62:7000",
						},
					},
				},
			},
			want: []*pb.AssignmentGroup{
				{
					Assignment: &pb.Assignment{
						Connection: "66.211.39.62:7000",
					},
				},
				{
					Assignment: &pb.Assignment{
						Connection: "45.211.39.62:7000",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := CleanUpAssignmentsWithoutConnection(tc.args.group)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestAssignTickets_assignTickets(t *testing.T) {
	type want struct {
		assigned  int
		wantError bool
		err       error
	}

	testCases := []struct {
		name    string
		request *pb.AssignTicketsRequest
		want    want
	}{
		{
			name: "it should return error for assignment without connection set",
			request: &pb.AssignTicketsRequest{
				Assignments: []*pb.AssignmentGroup{
					{
						Assignment: &pb.Assignment{},
					},
				},
			},
			want: want{
				assigned:  0,
				wantError: true,
				err:       errors.New("the AssignTicketsRequest does not have assignments with connections set"),
			},
		},
		{
			name: "it should assign 2 Assignments for a request with 2 connections set and 1 not set",
			request: &pb.AssignTicketsRequest{
				Assignments: []*pb.AssignmentGroup{
					{
						Assignment: &pb.Assignment{},
					},
					{
						Assignment: &pb.Assignment{
							Connection: "66.211.39.62:7000",
						},
					},
					{
						Assignment: &pb.Assignment{
							Connection: "45.211.39.62:8000",
						},
					},
				},
			},
			want: want{
				assigned:  2,
				wantError: false,
				err:       nil,
			},
		},
		{
			name: "it should assign 1 Assignment for a request with 1 connection set",
			request: &pb.AssignTicketsRequest{
				Assignments: []*pb.AssignmentGroup{
					{
						Assignment: &pb.Assignment{
							Connection: "66.211.39.62:7000",
						},
					},
				},
			},
			want: want{
				assigned:  1,
				wantError: false,
				err:       nil,
			},
		},
		{
			name: "it should assign 2 Assignments for a request with 2 connections set",
			request: &pb.AssignTicketsRequest{
				Assignments: []*pb.AssignmentGroup{
					{
						Assignment: &pb.Assignment{
							Connection: "66.211.39.62:7000",
						},
					},
					{
						Assignment: &pb.Assignment{
							Connection: "45.211.39.62:8000",
						},
					},
				},
			},
			want: want{
				assigned:  2,
				wantError: false,
				err:       nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assigner := &mockAssigner{}
			ctx := context.Background()
			assigner.On("AssignTickets", ctx, tc.request).Return(&pb.AssignTicketsResponse{}, tc.want.err)
			got, err := assignTickets(context.Background(), tc.request, assigner)

			if tc.want.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assigner.AssertExpectations(t)
				require.Equal(t, tc.want.assigned, got)
			}
		})
	}
}

type mockAssigner struct {
	mock.Mock
}

func (m *mockAssigner) AssignTickets(ctx context.Context, in *pb.AssignTicketsRequest, opts ...grpc.CallOption) (*pb.AssignTicketsResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.AssignTicketsResponse), args.Error(1)
}
