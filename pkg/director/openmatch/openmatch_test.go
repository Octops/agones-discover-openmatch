package openmatch

import (
	"github.com/stretchr/testify/require"
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
