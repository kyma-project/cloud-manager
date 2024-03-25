package util

import (
	"go.uber.org/zap/zapcore"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func GcpZapEncoderConfigOption() zap.EncoderConfigOption {
	return func(config *zapcore.EncoderConfig) {
		config.TimeKey = "time"
		config.LevelKey = "severity"
		config.NameKey = "logger"
		config.CallerKey = "caller"
		config.MessageKey = "message"
		config.StacktraceKey = "stacktrace"
		config.LineEnding = zapcore.DefaultLineEnding
		config.EncodeLevel = gcpLogLevelEncoder()
		config.EncodeTime = zapcore.RFC3339NanoTimeEncoder
		config.EncodeDuration = zapcore.MillisDurationEncoder
		config.EncodeCaller = zapcore.ShortCallerEncoder
	}
}

func gcpLogLevelEncoder() zapcore.LevelEncoder {
	return func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		switch l {
		case zapcore.DebugLevel:
			enc.AppendString("DEBUG")
		case zapcore.InfoLevel:
			enc.AppendString("INFO")
		case zapcore.WarnLevel:
			enc.AppendString("WARNING")
		case zapcore.ErrorLevel:
			enc.AppendString("ERROR")
		case zapcore.DPanicLevel:
			enc.AppendString("CRITICAL")
		case zapcore.PanicLevel:
			enc.AppendString("ALERT")
		case zapcore.FatalLevel:
			enc.AppendString("EMERGENCY")
		default:
			enc.AppendString(l.String())
		}
	}
}
