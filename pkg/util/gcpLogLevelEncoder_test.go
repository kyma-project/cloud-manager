package util

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestGcpLogLevelEncoderMapsEachLevelToGcpSeverity(t *testing.T) {
	cfg := zap.NewProductionEncoderConfig()
	GcpZapEncoderConfigOption()(&cfg)
	encoder := zapcore.NewJSONEncoder(cfg)

	cases := []struct {
		level    zapcore.Level
		severity string
	}{
		{zapcore.DebugLevel, "DEBUG"},
		{zapcore.InfoLevel, "INFO"},
		{zapcore.WarnLevel, "WARNING"},
		{zapcore.ErrorLevel, "ERROR"},
		{zapcore.DPanicLevel, "CRITICAL"},
		{zapcore.PanicLevel, "ALERT"},
		{zapcore.FatalLevel, "EMERGENCY"},
		// default branch: an unmapped level passes through as zap's own string.
		{zapcore.Level(-2), zapcore.Level(-2).String()},
	}

	for _, tc := range cases {
		t.Run(tc.severity, func(t *testing.T) {
			buf, err := encoder.EncodeEntry(zapcore.Entry{Level: tc.level, Message: "m"}, nil)
			require.NoError(t, err)

			var out map[string]any
			require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
			require.Equal(t, tc.severity, out["severity"])
		})
	}
}
