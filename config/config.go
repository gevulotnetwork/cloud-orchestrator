package config

import (
	"log"
	"sync"

	"github.com/nanovms/ops/cmd"
	api "github.com/nanovms/ops/lepton"
	"github.com/nanovms/ops/types"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var readOnce sync.Once

func readConfig() {
	viper.SetConfigType("toml")
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/gevulot")
	viper.AddConfigPath("$HOME/.config/gevulot")
	viper.AddConfigPath("$HOME/.gevulot")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	// Read config only once.
	readOnce.Do(readConfig)
}

type Factory struct {
	defaults             *cmd.MergeConfigContainer
	platformConfigurator PlatformConfig
}

// Platform defines the cloud provider, currently supporting aws, azure, and gcp.
func NewFactory() *Factory {
	platformCfg, err := getPlatformConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Construct all default config handlers.
	flags := newFlags()

	// Declare flags.
	cmd.PersistBuildImageCommandFlags(flags)
	cmd.PersistConfigCommandFlags(flags)
	cmd.PersistGlobalCommandFlags(flags)
	cmd.PersistNanosVersionCommandFlags(flags)
	cmd.PersistNightlyCommandFlags(flags)
	cmd.PersistPkgCommandFlags(flags)
	cmd.PersistProviderCommandFlags(flags)

	// Set some platform specific settings.
	platformCfg.SetFlags(flags)

	// Initialize config containers.
	configFlags := cmd.NewConfigCommandFlags(flags)
	globalFlags := cmd.NewGlobalCommandFlags(flags)
	nightlyFlags := cmd.NewNightlyCommandFlags(flags)
	nanosVersionFlags := cmd.NewNanosVersionCommandFlags(flags)
	buildImageFlags := cmd.NewBuildImageCommandFlags(flags)
	providerFlags := cmd.NewProviderCommandFlags(flags)
	pkgFlags := cmd.NewPkgCommandFlags(flags)

	return &Factory{
		defaults:             cmd.NewMergeConfigContainer(configFlags, globalFlags, nightlyFlags, nanosVersionFlags, buildImageFlags, providerFlags, pkgFlags),
		platformConfigurator: platformCfg,
	}
}

func getPlatformConfig() (PlatformConfig, error) {
	var platform PlatformConfig

	switch viper.GetString("platform") {
	case "aws":
		platform = &AWSPlatformConfig{}
	case "gcp":
		platform = &GCPPlatformConfig{}
	case "azure":
		log.Fatal("Azure is not supported at the moment")
	}

	err := platform.load()
	if err != nil {
		return nil, err
	}

	return platform, nil
}

// TODO: program param is reserved for program specific configuration lookups.
func (f *Factory) NewConfig(program string) *types.Config {
	cfg := api.NewConfig()
	err := f.defaults.Merge(cfg)
	if err != nil {
		// There are no user passed parameters involved here so there should be no
		// way for `Merge(cfg)` to return an error here.
		log.Fatalf("failed to default config: %#v", err)
	}

	// Finalize last minute adjustments to config ;)
	f.platformConfigurator.FinalizeConfig(cfg)

	return cfg
}

func newFlags() *pflag.FlagSet {
	// TODO: Figure out if flags.SetOutput(new(bytes.Buffer)) is needed.
	return pflag.NewFlagSet("<placeholder>", pflag.ContinueOnError)
}
