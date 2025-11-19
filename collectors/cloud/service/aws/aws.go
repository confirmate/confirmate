package aws

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"confirmate.io/collectors/cloud/internal/logconfig"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"
)

var (
	log *slog.Logger

	// loadDefaultConfig holds config.LoadDefaultConfig() so that NewClient() can use it and test function can mock it
	loadDefaultConfig = config.LoadDefaultConfig

	// newFromConfigSTS holds sts.NewFromConfig() so that NewClient() can use it and test function can mock it
	newFromConfigSTS = loadSTSClient
)

// Client holds configurations across all services within AWS
type Client struct {
	// cfg holds AWS SDK configuration
	cfg aws.Config

	// accountID is needed for ARN creation
	accountID *string
}

// STSAPI describes the STS api interface which is implemented by the official AWS client and mock clients in tests
type STSAPI interface {
	GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

func init() {
	log = logconfig.GetLogger().With("component", "aws-discovery")
}

// NewClient constructs a new AwsClient
// TODO(lebogg): "Overload" (switch) with staticCredentialsProvider
func NewClient() (*Client, error) {
	c := &Client{}

	// load configuration
	cfg, err := loadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("could not load default config: %w", err)
	}
	c.cfg = cfg

	// load accountID
	stsClient := newFromConfigSTS(cfg)
	resp, err := stsClient.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, prettyError(err)
	}
	c.accountID = resp.Account

	return c, err
}

// formatError returns AWS API specific error code transformed into the default error type
func formatError(ae smithy.APIError) error {
	return fmt.Errorf("code: %v, fault: %v, message: %v", ae.ErrorCode(), ae.ErrorFault(), ae.ErrorMessage())
}

// prettyError returns an AWS API specific error code if it is an AWS error (using [formatError]), otherwise, just the error itself.
func prettyError(err error) error {
	var ae smithy.APIError
	if errors.As(err, &ae) {
		err = formatError(ae)
	}
	return err
}

// loadSTSClient creates the STS client using the STS api interface (for mock testing)
func loadSTSClient(cfg aws.Config) STSAPI {
	client := sts.NewFromConfig(cfg)
	return client
}
