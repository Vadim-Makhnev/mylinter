package analyzer

import "golang.org/x/tools/go/analysis"

type Settings struct {
	ConfigPath string
}

var cliSettings Settings

var Analyzer = newAnalyzer(func() Settings {
	return cliSettings
})

func init() {
	Analyzer.Flags.StringVar(
		&cliSettings.ConfigPath,
		"config",
		"",
		"path to mylinter config file",
	)
}

func NewAnalyzer(settings Settings) *analysis.Analyzer {
	return newAnalyzer(func() Settings {
		return settings
	})
}

func newAnalyzer(settingsProvider func() Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "mylinter",
		Doc:  "log linter",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, settingsProvider())
		},
	}
}
