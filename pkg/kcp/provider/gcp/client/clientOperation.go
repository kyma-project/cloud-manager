package client

import (
	"context"

	filestore "cloud.google.com/go/filestore/apiv1"
	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/protobuf/protoadapt"
)

type BaseOperation interface {
	Name() string
	Done() bool
}

type ResultOperation[T protoadapt.MessageV1] interface {
	BaseOperation
	Poll(ctx context.Context, opts ...gax.CallOption) (T, error)
	Wait(ctx context.Context, opts ...gax.CallOption) (T, error)
}

var _ ResultOperation[*filestorepb.Instance] = (*filestore.CreateInstanceOperation)(nil)

type VoidOperation interface {
	BaseOperation
	Poll(ctx context.Context, opts ...gax.CallOption) error
	Wait(ctx context.Context, opts ...gax.CallOption) error
}

var _ VoidOperation = (*filestore.DeleteInstanceOperation)(nil)
