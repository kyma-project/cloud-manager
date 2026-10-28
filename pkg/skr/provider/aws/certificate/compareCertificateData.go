package certificate

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func compareCertificateData(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.certificateDetail == nil {
		logger.Info("Certificate not found in ACM, import needed")
		state.certificateNeedsUpdate = true
		return nil, ctx
	}

	secretSerial := formatSerial(state.certificateData.X509.SerialNumber.Bytes())
	acmSerial := ptr.Deref(state.certificateDetail.Serial, "")

	if secretSerial != acmSerial {
		logger.Info("Certificate serial differs, update needed",
			"secretSerial", secretSerial,
			"acmSerial", acmSerial,
		)
		state.certificateNeedsUpdate = true
		return nil, ctx
	}

	state.certificateNeedsUpdate = false
	return nil, ctx
}
