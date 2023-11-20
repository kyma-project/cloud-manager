package composed

import (
	"go.uber.org/zap"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type BaseState struct {
	logger *zap.SugaredLogger
	client.Client
	record.EventRecorder
}

func (me *BaseState) Logger() *zap.SugaredLogger {
	return me.logger
}
