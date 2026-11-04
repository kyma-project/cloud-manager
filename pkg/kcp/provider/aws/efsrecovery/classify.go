package efsrecovery

import (
	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/kyma-project/cloud-manager/pkg/common"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	"k8s.io/utils/ptr"
)

// classification is the outcome of grouping in-scope EFS against the KCP operational truth.
type classification struct {
	keepers   []efstypes.FileSystemDescription
	orphans   []efstypes.FileSystemDescription
	ambiguous []efstypes.FileSystemDescription
}

// efsResourceName derives the owning NfsInstance name of an already-in-scope EFS: the
// cloud-manager name tag if present, else the Name tag. Used only to group/log in-scope EFS —
// never to widen scope.
func efsResourceName(fs *efstypes.FileSystemDescription) string {
	if v := awsutil.GetEfsTagValue(fs.Tags, common.TagCloudManagerName); v != "" {
		return v
	}
	return awsutil.GetEfsTagValue(fs.Tags, "Name")
}

// inScope keeps only EFS whose cloud-manager shoot tag equals the configured shoot id. There is
// deliberately no Name-tag fallback here: an EFS with no cloud-manager shoot tag is out of scope,
// full stop. This is the absolute protection for the healthy fleet.
func inScope(all []efstypes.FileSystemDescription, shootId string) []efstypes.FileSystemDescription {
	if shootId == "" {
		return nil
	}
	var out []efstypes.FileSystemDescription
	for i := range all {
		if awsutil.GetEfsTagValue(all[i].Tags, common.TagShoot) == shootId {
			out = append(out, all[i])
		}
	}
	return out
}

// classify joins in-scope EFS against the set of non-empty NfsInstance status.id values.
//   - keeper  = FileSystemId is some NfsInstance's status.id
//   - orphan  = in-scope but not a status.id
//   - ambiguous = defensive: should not arise under the status.id join, but if an EFS with mount
//     targets is not a keeper and cannot be attributed, keep it and log rather than delete.
func classify(scoped []efstypes.FileSystemDescription, statusIds map[string]struct{}, mountTargetCounts map[string]int) classification {
	var c classification
	for i := range scoped {
		fs := scoped[i]
		id := ptr.Deref(fs.FileSystemId, "")
		if _, isKeeper := statusIds[id]; isKeeper {
			c.keepers = append(c.keepers, fs)
			continue
		}
		// not a keeper. If it has mount targets but we cannot attribute it to any resource name,
		// treat as ambiguous (keep + log) instead of deleting.
		if mountTargetCounts[id] > 0 && efsResourceName(&fs) == "" {
			c.ambiguous = append(c.ambiguous, fs)
			continue
		}
		c.orphans = append(c.orphans, fs)
	}
	return c
}
