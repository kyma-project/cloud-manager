package efsrecovery

import (
	"testing"

	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

func mkEfs(id string, tags map[string]string, mountTargets int32) efstypes.FileSystemDescription {
	var t []efstypes.Tag
	for k, v := range tags {
		t = append(t, efstypes.Tag{Key: new(k), Value: new(v)})
	}
	return efstypes.FileSystemDescription{
		FileSystemId:         new(id),
		Tags:                 t,
		NumberOfMountTargets: mountTargets,
	}
}

func TestInScope(t *testing.T) {
	const shoot = "shoot-a"

	all := []efstypes.FileSystemDescription{
		mkEfs("in-1", map[string]string{common.TagShoot: shoot}, 0),
		mkEfs("other-shoot", map[string]string{common.TagShoot: "shoot-b"}, 0),
		mkEfs("no-shoot-tag", map[string]string{"Name": shoot}, 0), // Name tag must NOT widen scope
		mkEfs("untagged", nil, 0),
		mkEfs("in-2", map[string]string{common.TagShoot: shoot, common.TagCloudManagerName: "r2"}, 0),
	}

	scoped := inScope(all, shoot)

	ids := map[string]bool{}
	for _, fs := range scoped {
		ids[ptr.Deref(fs.FileSystemId, "")] = true
	}
	assert.Len(t, scoped, 2)
	assert.True(t, ids["in-1"])
	assert.True(t, ids["in-2"])
	assert.False(t, ids["no-shoot-tag"], "Name tag must never bring an EFS into scope")
	assert.False(t, ids["other-shoot"])
	assert.False(t, ids["untagged"])

	t.Run("empty shoot id scopes nothing", func(t *testing.T) {
		assert.Empty(t, inScope(all, ""))
	})
}

func TestClassify(t *testing.T) {
	const shoot = "shoot-a"
	scoped := []efstypes.FileSystemDescription{
		mkEfs("keeper-1", map[string]string{common.TagShoot: shoot, common.TagCloudManagerName: "r1"}, 1),
		mkEfs("orphan-dup", map[string]string{common.TagShoot: shoot, common.TagCloudManagerName: "r1"}, 0),
		mkEfs("orphan-notready", map[string]string{common.TagShoot: shoot, common.TagCloudManagerName: "r2"}, 0),
		mkEfs("ambiguous", map[string]string{common.TagShoot: shoot}, 2), // has MTs, no attributable name
	}
	statusIds := map[string]struct{}{"keeper-1": {}}
	mtCounts := countMountTargets(scoped)

	c := classify(scoped, statusIds, mtCounts)

	assert.Equal(t, []string{"keeper-1"}, idsOf(c.keepers))
	assert.ElementsMatch(t, []string{"orphan-dup", "orphan-notready"}, idsOf(c.orphans))
	assert.Equal(t, []string{"ambiguous"}, idsOf(c.ambiguous))
}

func TestClassify_NameFallbackGroupsButDoesNotRescueFromOrphan(t *testing.T) {
	const shoot = "shoot-a"
	// in-scope EFS attributable only by Name tag, not a status.id → still an orphan (not ambiguous),
	// because it has a derivable resource name even without the cloud-manager name tag.
	scoped := []efstypes.FileSystemDescription{
		mkEfs("orphan-byname", map[string]string{common.TagShoot: shoot, "Name": "r3"}, 1),
	}
	c := classify(scoped, map[string]struct{}{}, countMountTargets(scoped))
	assert.Equal(t, []string{"orphan-byname"}, idsOf(c.orphans))
	assert.Empty(t, c.ambiguous)
	assert.Empty(t, c.keepers)
}

func idsOf(list []efstypes.FileSystemDescription) []string {
	out := make([]string, 0, len(list))
	for i := range list {
		out = append(out, ptr.Deref(list[i].FileSystemId, ""))
	}
	return out
}
