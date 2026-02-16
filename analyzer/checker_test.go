package analyzer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	td := analysistest.TestData()

	cases := []string{
		"a",
	}

	for _, pkg := range cases {
		analysistest.Run(t, td, Analyzer, pkg)
	}
}

func TestChecker_IsEnglishOnly(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want bool
	}{
		{
			name: "normal string",
			str:  "hello",
			want: true,
		},
		{
			name: "with space between",
			str:  "hello world",
			want: true,
		},
		{
			name: "upper case letter",
			str:  "Hello",
			want: true,
		},
		{
			name: "empty string",
			str:  "",
			want: false,
		},
		{
			name: "ciryllic string",
			str:  "Привет",
			want: false,
		},
		{
			name: "spec symbols",
			str:  "%#$!%^@&*#(@&)",
			want: false,
		},
		{
			name: "mix string",
			str:  "hello мир",
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			ok := isEnglishOnly(tc.str)

			assert.Equal(t, tc.want, ok)
		})
	}
}

func TestChecker_IsFirstLower(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want bool
	}{
		{
			name: "normal string",
			str:  "hello",
			want: true,
		},
		{
			name: "upper case letter",
			str:  "Hello",
			want: false,
		},
		{
			name: "empty string",
			str:  "",
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			ok := isFirstLower(tc.str)

			assert.Equal(t, tc.want, ok)
		})
	}
}

func TestChecker_NormalizeSensitiveInput(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want string
	}{
		{
			name: "normal string",
			str:  "hello world",
			want: "helloworld",
		},
		{
			name: "empty string",
			str:  " ",
			want: "",
		},
		{
			name: "three spaces",
			str:  "   ",
			want: "",
		},
		{
			name: "mixed case and separators",
			str:  "API_Key: Client-Secret.",
			want: "apikeyclientsecret",
		},
		{
			name: "dots dashes and underscores",
			str:  "my.private-key_value",
			want: "myprivatekeyvalue",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			ok := normalizeSensitiveInput(tc.str)

			assert.Equal(t, tc.want, ok)
		})
	}
}

func TestChecker_ContainsSensitiveKeyword(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want bool
	}{
		{
			name: "normal string",
			str:  "password pwd passwd",
			want: true,
		},
		{
			name: "empty string",
			str:  " ",
			want: false,
		},
		{
			name: "wrong string",
			str:  "hello world",
			want: false,
		},
		{
			name: "api key with underscore",
			str:  "api_key=abc",
			want: true,
		},
		{
			name: "client secret with dash",
			str:  "Client-Secret: value",
			want: true,
		},
		{
			name: "authorization header",
			str:  "Authorization: Bearer token",
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			ok := containsSensitiveKeyword(tc.str)

			assert.Equal(t, tc.want, ok)
		})
	}
}

func TestChecker_ResolveSensitiveKeywords_FromConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "mylinter.yml")
	configData := []byte(`
sensitive_keywords:
  add:
    - session_id
    - refresh-token
  remove:
    - token
    - auth
`)

	err := os.WriteFile(configPath, configData, 0o644)
	assert.NoError(t, err)

	got, err := resolveSensitiveKeywords(configPath)
	assert.NoError(t, err)

	assert.Contains(t, got, "password")
	assert.Contains(t, got, "sessionid")
	assert.Contains(t, got, "refreshtoken")
	assert.NotContains(t, got, "token")
	assert.NotContains(t, got, "auth")
}

func TestChecker_ResolveSensitiveKeywords_ValuesOverride(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "mylinter.yml")
	configData := []byte(`
sensitive_keywords:
  values:
    - custom_key
    - token
  add:
    - one-time-secret
  remove:
    - token
`)

	err := os.WriteFile(configPath, configData, 0o644)
	assert.NoError(t, err)

	got, err := resolveSensitiveKeywords(configPath)
	assert.NoError(t, err)

	assert.Contains(t, got, "customkey")
	assert.Contains(t, got, "onetimesecret")
	assert.NotContains(t, got, "password")
	assert.NotContains(t, got, "token")
}
