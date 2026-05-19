package feature

import "context"

const (
	runtimeSecurityAwsFlagName   = "runtimeSecurityAws"
	runtimeSecurityAzureFlagName = "runtimeSecurityAzure"
	runtimeSecurityGcpFlagName   = "runtimeSecurityGcp"
)

var RuntimeSecurityAws = &runtimeSecurityInfo{flagName: runtimeSecurityAwsFlagName}
var RuntimeSecurityAzure = &runtimeSecurityInfo{flagName: runtimeSecurityAzureFlagName}
var RuntimeSecurityGcp = &runtimeSecurityInfo{flagName: runtimeSecurityGcpFlagName}

type runtimeSecurityInfo struct {
	flagName string
}

func (r *runtimeSecurityInfo) Value(ctx context.Context) bool {
	v := provider.BoolVariation(ctx, r.flagName, false)
	return v
}
