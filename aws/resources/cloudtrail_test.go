package resources

import (
	"context"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	awsgo "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/aws/aws-sdk-go/service/cloudtrail/cloudtrailiface"
	"github.com/gruntwork-io/cloud-nuke/config"
	"github.com/stretchr/testify/require"
)

type mockedCloudTrail struct {
	cloudtrailiface.CloudTrailAPI
	ListTrailsOutput  cloudtrail.ListTrailsOutput
	DeleteTrailOutput cloudtrail.DeleteTrailOutput
}

func (m mockedCloudTrail) ListTrailsPagesWithContext(_ awsgo.Context, _ *cloudtrail.ListTrailsInput, fn func(*cloudtrail.ListTrailsOutput, bool) bool, _ ...request.Option) error {
	fn(&m.ListTrailsOutput, true)
	return nil
}

func (m mockedCloudTrail) DeleteTrailWithContext(_ aws.Context, _ *cloudtrail.DeleteTrailInput, _ ...request.Option) (*cloudtrail.DeleteTrailOutput, error) {
	return &m.DeleteTrailOutput, nil
}

func TestCloudTrailGetAll(t *testing.T) {

	t.Parallel()

	testName1 := "test-name1"
	testName2 := "test-name2"
	testArn1 := "test-arn1"
	testArn2 := "test-arn2"
	ct := CloudtrailTrail{
		Client: mockedCloudTrail{
			ListTrailsOutput: cloudtrail.ListTrailsOutput{
				Trails: []*cloudtrail.TrailInfo{
					{
						Name:     aws.String(testName1),
						TrailARN: aws.String(testArn1),
					},
					{
						Name:     aws.String(testName2),
						TrailARN: aws.String(testArn2),
					},
				}}}}

	tests := map[string]struct {
		configObj config.ResourceType
		expected  []string
	}{
		"emptyFilter": {
			configObj: config.ResourceType{},
			expected:  []string{testArn1, testArn2},
		},
		"nameExclusionFilter": {
			configObj: config.ResourceType{
				ExcludeRule: config.FilterRule{
					NamesRegExp: []config.Expression{{
						RE: *regexp.MustCompile(testName1),
					}}},
			},
			expected: []string{testArn2},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			names, err := ct.getAll(context.Background(), config.Config{
				CloudtrailTrail: tc.configObj,
			})

			require.NoError(t, err)
			require.Equal(t, tc.expected, aws.StringValueSlice(names))
		})
	}
}

func TestCloudTrailNukeAll(t *testing.T) {

	t.Parallel()

	ct := CloudtrailTrail{
		Client: mockedCloudTrail{
			DeleteTrailOutput: cloudtrail.DeleteTrailOutput{},
		}}

	err := ct.nukeAll([]*string{aws.String("test-arn")})
	require.NoError(t, err)
}
