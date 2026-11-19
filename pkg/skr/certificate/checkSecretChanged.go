package certificate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// computeSecretHash returns a hex-encoded SHA-256 of the certificate secret's relevant keys.
//
// Raw bytes are hashed rather than using the X.509 certificate fingerprint or serial number because:
//   - A fingerprint covers only tls.crt; changes to tls.key or ca.crt would be missed.
//   - Serial number is set by the CA and does not change when the same cert is re-uploaded.
//   - Both require successfully parsing the certificate before the check can run; hashing raw
//     bytes works even when the content is invalid, letting the short-circuit fire consistently.
func computeSecretHash(state *State) string {
	h := sha256.New()
	h.Write(state.secret.Data["tls.crt"])
	h.Write(state.secret.Data["tls.key"])
	h.Write(state.secret.Data["ca.crt"])
	return hex.EncodeToString(h.Sum(nil))
}

// checkSecretChanged short-circuits the reconcile if the secret content has not changed
// and the certificate has already been imported (ProviderId is set).
// This avoids unnecessary AWS API calls on every reconcile regardless of the current status —
// if the secret content is unchanged, there is nothing new to act on.
func checkSecretChanged(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	cert := state.ObjAsCertificate()

	hash := computeSecretHash(state)
	state.secretHash = hash

	if cert.Status.SecretHash == hash && cert.Status.ProviderId != "" {
		composed.LoggerFromCtx(ctx).Info("Secret content unchanged, skipping AWS API calls")
		return composed.StopAndForget, ctx
	}

	return nil, ctx
}
