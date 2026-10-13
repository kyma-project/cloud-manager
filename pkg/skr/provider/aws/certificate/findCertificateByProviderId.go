package certificate

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func findCertificateByProviderId(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	cert := state.ObjAsCertificate()

	if cert.Status.ProviderId == "" {
		return nil, ctx
	}

	state.certificateArn = cert.Status.ProviderId
	return nil, ctx
}
