package mock
//
//import (
//	"context"
//
//	"github.com/aws/aws-sdk-go-v2/service/sts"
//	"k8s.io/utils/ptr"
//)
//
//type ScopeConfig interface {
//	SetAccount(string)
//	GetAccount() string
//}
//
//type scopeStore struct {
//	account string
//}
//
//func (s *scopeStore) SetAccount(account string) {
//	s.account = account
//}
//
//func (s *scopeStore) GetAccount() string {
//	return s.account
//}
//
//func (s *scopeStore) GetCallerIdentity(ctx context.Context) (*sts.GetCallerIdentityOutput, error) {
//	if isContextCanceled(ctx) {
//		return nil, context.Canceled
//	}
//	return &sts.GetCallerIdentityOutput{
//		Account: ptr.To(s.account),
//		Arn:     nil,
//		UserId:  nil,
//	}, nil
//}
