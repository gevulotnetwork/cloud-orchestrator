package config

import (
	"github.com/nanovms/ops/types"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type PlatformConfig interface {
	SetFlags(*pflag.FlagSet)
	FinalizeConfig(*types.Config)
	load() error
}

type GCPPlatformConfig struct {
	Bucket    string
	ProjectID string
	Zone      string
}

func (c *GCPPlatformConfig) SetFlags(flags *pflag.FlagSet) {
	// TODO: Apparently CloudConfig.Bucket doesn't have a flag?
	flags.Set("target-cloud", "gcp")
	flags.Set("projectid", c.ProjectID)
	flags.Set("zone", c.Zone)
}

func (c *GCPPlatformConfig) FinalizeConfig(cfg *types.Config) {
	// Bucket doesn't have FlagSet value for some reason, so it needs to be set here.
	cfg.CloudConfig.BucketName = c.Bucket
}

func (c *GCPPlatformConfig) load() error {
	c.Bucket = viper.GetString("gcp.bucket")
	c.ProjectID = viper.GetString("gcp.project_id")
	c.Zone = viper.GetString("gcp.zone")
	return nil
}

type AWSPlatformConfig struct {
}

func (c *AWSPlatformConfig) SetFlags(flags *pflag.FlagSet) {
	flags.Set("target-cloud", "aws")
}

func (c *AWSPlatformConfig) FinalizeConfig(cfg *types.Config) {
	// NOP
}

func (c *AWSPlatformConfig) load() error {
	return nil
}
