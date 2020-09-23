package config

import (
	"context"
	"fmt"

	"github.com/grupawp/tensorflow-deploy/app"
	"github.com/grupawp/tensorflow-deploy/config/structmerge"
	"github.com/grupawp/tensorflow-deploy/exterr"
	"github.com/grupawp/tensorflow-deploy/logging"
)

type Config struct {
	defaults *app.Config
	cli      *app.Config
	env      *app.Config
	yamlFile *app.Config
}

// NewConfig ...
func NewConfig(ctx context.Context) (*Config, error) {
	var err error
	c := &Config{}

	// sources
	c.defaults, err = app.NewConfigDefaults(ctx)
	if err != nil {
		return nil, exterr.WrapWithFrame(err)
	}

	c.cli, err = app.NewConfigCLI(ctx)
	if err != nil {
		if err == app.ErrCLIUsage {
			return nil, err
		}

		return nil, exterr.WrapWithFrame(err)
	}

	c.env, err = app.NewConfigENV(ctx, "")
	if err != nil {
		return nil, exterr.WrapWithFrame(err)
	}

	var yamlFile string
	if c.cli.App.ConfigFile != nil {
		yamlFile = *c.cli.App.ConfigFile
	} else if c.env.App.ConfigFile != nil {
		yamlFile = *c.env.App.ConfigFile
	}
	c.yamlFile, err = app.NewConfigYAML(ctx, yamlFile)
	if err != nil {
		return nil, exterr.WrapWithFrame(err)
	}

	return c, nil
}

// Parse ...
func (c *Config) Parse(ctx context.Context) (*app.Config, error) {
	configParams := app.NewConfig()
	sm := structmerge.NewStructMerge().WithSource("Defaults", c.defaults).WithSource("CLI", c.cli).WithSource("ENV", c.env).WithSource("YAML-FILE", c.yamlFile)
	if err := sm.Merge(configParams); err != nil {
		return nil, exterr.WrapWithFrame(err)
	}

	for _, v := range sm.MergedMeta {
		logging.Info(ctx, fmt.Sprintf("Config parameter `%s` set from `%s`", v.FieldPath, v.Source))
	}

	if err := configParams.Validate(ctx); err != nil {
		return nil, exterr.WrapWithFrame(err)
	}

	return configParams, nil
}
