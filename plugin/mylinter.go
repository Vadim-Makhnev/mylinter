package main

import (
	"fmt"

	"github.com/Vadim-Makhnev/mylinter/analyzer"
	"golang.org/x/tools/go/analysis"
)

type pluginSettings struct {
	Config string
}

func New(conf any) ([]*analysis.Analyzer, error) {
	settings, err := decodePluginSettings(conf)
	if err != nil {
		return nil, err
	}

	return []*analysis.Analyzer{
		analyzer.NewAnalyzer(analyzer.Settings{
			ConfigPath: settings.Config,
		}),
	}, nil
}

func decodePluginSettings(conf any) (pluginSettings, error) {
	if conf == nil {
		return pluginSettings{}, nil
	}

	values, ok := conf.(map[string]any)
	if !ok {
		return pluginSettings{}, fmt.Errorf("invalid plugin settings type: %T", conf)
	}

	var settings pluginSettings

	if rawConfig, ok := values["config"]; ok {
		configPath, ok := rawConfig.(string)
		if !ok {
			return pluginSettings{}, fmt.Errorf("settings.config must be a string, got %T", rawConfig)
		}
		settings.Config = configPath
	}

	return settings, nil
}
