package alicloud

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
)

const apsaraDBCACertURL = "https://apsaradb-public.oss-ap-southeast-1.aliyuncs.com/ApsaraDB-CA-Chain.zip"

// FetchApsaraDBCACert downloads the AliCloud ApsaraDB CA certificate chain
// from the public URL and returns the PEM content. AliCloud r-kvstore uses a
// proprietary CA that is not in the standard system trust store.
func FetchApsaraDBCACert(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apsaraDBCACertURL, nil)
	if err != nil {
		return "", fmt.Errorf("building ApsaraDB CA request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching ApsaraDB CA chain: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ApsaraDB CA chain fetch returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return "", fmt.Errorf("reading ApsaraDB CA chain response: %w", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return "", fmt.Errorf("opening ApsaraDB CA chain zip: %w", err)
	}

	for _, f := range zr.File {
		if strings.HasSuffix(f.Name, ".pem") {
			rc, err := f.Open()
			if err != nil {
				return "", fmt.Errorf("opening %s in ApsaraDB CA zip: %w", f.Name, err)
			}
			defer rc.Close()
			pem, err := io.ReadAll(rc)
			if err != nil {
				return "", fmt.Errorf("reading %s from ApsaraDB CA zip: %w", f.Name, err)
			}
			return string(pem), nil
		}
	}
	return "", fmt.Errorf("no .pem file found in ApsaraDB CA chain zip")
}

// GeneratePassword returns a 32-char password satisfying AliCloud r-kvstore
// requirements: at least one uppercase, one lowercase, one digit.
func GeneratePassword() string {
	const upper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const lower = "abcdefghijklmnopqrstuvwxyz"
	const digits = "0123456789"
	const all = upper + lower + digits

	randChar := func(charset string) byte {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			panic(fmt.Sprintf("crypto/rand failure: %v", err))
		}
		return charset[n.Int64()]
	}

	b := make([]byte, 32)
	b[0] = randChar(upper)
	b[1] = randChar(lower)
	b[2] = randChar(digits)
	for i := 3; i < 32; i++ {
		b[i] = randChar(all)
	}
	for i := len(b) - 1; i > 0; i-- {
		jBig, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			panic(fmt.Sprintf("crypto/rand failure: %v", err))
		}
		j := int(jBig.Int64())
		b[i], b[j] = b[j], b[i]
	}
	return string(b)
}

// BuildRequiredCidrs returns the unique list of CIDRs that must be present in the security-IP list.
func BuildRequiredCidrs(nodesCidr, ipRangeCidr string) []string {
	required := []string{}
	if nodesCidr != "" {
		required = append(required, nodesCidr)
	}
	if ipRangeCidr != "" && ipRangeCidr != nodesCidr {
		required = append(required, ipRangeCidr)
	}
	return required
}

// HasAllCidrs reports whether all required CIDRs appear in the comma-separated existing string.
func HasAllCidrs(existing string, required []string) bool {
	parts := strings.Split(existing, ",")
	existingSet := make(map[string]struct{}, len(parts))
	for _, p := range parts {
		existingSet[strings.TrimSpace(p)] = struct{}{}
	}
	for _, r := range required {
		if _, ok := existingSet[r]; !ok {
			return false
		}
	}
	return true
}
