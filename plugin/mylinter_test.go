package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodePluginSettings(t *testing.T) {
	tests := []struct {
		name    string
		conf    any
		want    pluginSettings
		wantErr bool
	}{
		{
			name: "nil config",
			conf: nil,
			want: pluginSettings{},
		},
		{
			name: "valid config path",
			conf: map[string]any{
				"config": ".mylinter.yml",
			},
			want: pluginSettings{Config: ".mylinter.yml"},
		},
		{
			name: "wrong config type",
			conf: map[string]any{
				"config": 10,
			},
			wantErr: true,
		},
		{
			name:    "wrong settings type",
			conf:    "config",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := decodePluginSettings(tc.conf)

			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
