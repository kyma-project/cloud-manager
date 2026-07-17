//go:build integration

package client

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	vpc "github.com/alibabacloud-go/vpc-20160428/v6/client"
)

// These tests exercise the REAL AliCloud NAS API. They are gated behind the
// `integration` build tag and skip unless ALICLOUD credentials are present.
//
//	ALICLOUD_ACCESS_KEY=... ALICLOUD_SECRET_KEY=... \
//	  go test -tags integration ./pkg/kcp/provider/alicloud/nfsinstance/client/...
//
// NAS mount targets require a VPC and a vSwitch, so the setup provisions a
// temp VPC + vSwitch (and cleans them up), mirroring the iprange integration test.

const testRegion = "eu-central-1"

func newTestClient(t *testing.T) Client {
	t.Helper()
	accessKey := os.Getenv("ALICLOUD_ACCESS_KEY")
	secretKey := os.Getenv("ALICLOUD_SECRET_KEY")

	if accessKey == "" || secretKey == "" {
		t.Skip("ALICLOUD_ACCESS_KEY and ALICLOUD_SECRET_KEY must be set")
	}

	provider := NewClientProvider()
	client, err := provider(context.Background(), testRegion, accessKey, secretKey)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	return client
}

func newRawVpcClient(t *testing.T) *vpc.Client {
	t.Helper()
	config := &openapi.Config{
		AccessKeyId:     tea.String(os.Getenv("ALICLOUD_ACCESS_KEY")),
		AccessKeySecret: tea.String(os.Getenv("ALICLOUD_SECRET_KEY")),
		RegionId:        tea.String(testRegion),
	}
	config.Endpoint = tea.String(fmt.Sprintf("vpc.%s.aliyuncs.com", testRegion))
	vpcClient, err := vpc.NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create raw vpc client: %v", err)
	}
	return vpcClient
}

// createTempVpcWithVSwitch creates a VPC and a vSwitch in the first available zone.
// Returns (vpcId, vSwitchId, zoneId). Caller must call the returned cleanup func.
func createTempVpcWithVSwitch(t *testing.T) (string, string, string, func()) {
	t.Helper()
	vpcClient := newRawVpcClient(t)

	name := fmt.Sprintf("cm-nfs-test-%d", time.Now().UnixMilli())
	vpcResp, err := vpcClient.CreateVpc(&vpc.CreateVpcRequest{
		RegionId:  tea.String(testRegion),
		VpcName:   tea.String(name),
		CidrBlock: tea.String("10.0.0.0/8"),
	})
	if err != nil {
		t.Fatalf("Failed to create temp VPC: %v", err)
	}
	vpcId := tea.StringValue(vpcResp.Body.VpcId)
	t.Logf("Created temp VPC: %s (%s)", vpcId, name)
	time.Sleep(5 * time.Second) // wait for VPC to become Available

	zonesResp, err := vpcClient.DescribeZones(&vpc.DescribeZonesRequest{RegionId: tea.String(testRegion)})
	if err != nil {
		t.Fatalf("DescribeZones failed: %v", err)
	}
	if zonesResp.Body == nil || zonesResp.Body.Zones == nil || len(zonesResp.Body.Zones.Zone) == 0 {
		t.Fatal("no zones available")
	}
	zoneId := tea.StringValue(zonesResp.Body.Zones.Zone[0].ZoneId)

	vswResp, err := vpcClient.CreateVSwitch(&vpc.CreateVSwitchRequest{
		RegionId:    tea.String(testRegion),
		VpcId:       tea.String(vpcId),
		ZoneId:      tea.String(zoneId),
		CidrBlock:   tea.String("10.99.0.0/22"),
		VSwitchName: tea.String(name),
	})
	if err != nil {
		// best-effort VPC cleanup before failing
		_, _ = vpcClient.DeleteVpc(&vpc.DeleteVpcRequest{RegionId: tea.String(testRegion), VpcId: tea.String(vpcId)})
		t.Fatalf("Failed to create temp vSwitch: %v", err)
	}
	vSwitchId := tea.StringValue(vswResp.Body.VSwitchId)
	t.Logf("Created temp vSwitch: %s zone=%s", vSwitchId, zoneId)
	time.Sleep(3 * time.Second) // wait for vSwitch to become Available

	cleanup := func() {
		_, err := vpcClient.DeleteVSwitch(&vpc.DeleteVSwitchRequest{RegionId: tea.String(testRegion), VSwitchId: tea.String(vSwitchId)})
		if err != nil {
			t.Logf("Warning: failed to delete temp vSwitch %s: %v", vSwitchId, err)
		}
		time.Sleep(3 * time.Second)
		_, err = vpcClient.DeleteVpc(&vpc.DeleteVpcRequest{RegionId: tea.String(testRegion), VpcId: tea.String(vpcId)})
		if err != nil {
			t.Logf("Warning: failed to delete temp VPC %s: %v", vpcId, err)
		} else {
			t.Logf("Deleted temp VPC: %s", vpcId)
		}
	}

	return vpcId, vSwitchId, zoneId, cleanup
}

// pollFileSystemStatus waits until the file system reaches the desired status or times out.
func pollFileSystemStatus(t *testing.T, client Client, ctx context.Context, fsId, want string, timeout time.Duration) *FileSystemInfo {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		fs, err := client.DescribeFileSystem(ctx, fsId)
		if err != nil {
			t.Fatalf("DescribeFileSystem failed: %v", err)
		}
		if fs != nil && fs.Status == want {
			return fs
		}
		time.Sleep(3 * time.Second)
	}
	t.Fatalf("file system %s did not reach status %q within %s", fsId, want, timeout)
	return nil
}

func TestAlicloudNasFileSystemLifecycle(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Create file system
	t.Log("Creating NAS file system ...")
	fsId, err := client.CreateFileSystem(ctx, "NFS", "Performance", "")
	if err != nil {
		t.Fatalf("CreateFileSystem failed: %v", err)
	}
	t.Logf("File system created: %s", fsId)

	defer func() {
		t.Logf("Deleting file system %s ...", fsId)
		if err := client.DeleteFileSystem(ctx, fsId); err != nil {
			t.Logf("Warning: DeleteFileSystem failed: %v", err)
		}
	}()

	// Wait until Running
	fs := pollFileSystemStatus(t, client, ctx, fsId, "Running", 2*time.Minute)
	t.Logf("File system available: id=%s protocol=%s storage=%s zone=%s", fs.FileSystemId, fs.ProtocolType, fs.StorageType, fs.ZoneId)

	// Describe by id
	got, err := client.DescribeFileSystem(ctx, fsId)
	if err != nil {
		t.Fatalf("DescribeFileSystem failed: %v", err)
	}
	if got == nil || got.FileSystemId != fsId {
		t.Fatalf("DescribeFileSystem returned unexpected result: %+v", got)
	}
}

// TestAlicloudNasDescribeMissingResources probes the "describe a resource that does not
// exist" behavior for EVERY describe method. AliCloud NAS is inconsistent here — some
// describes return an empty result for an absent resource, others return HTTP 404. The
// reconciler loads-before-create, so every describe MUST surface "absent" as an empty
// result, not an error. This test asserts that contract against the REAL API for all four
// methods (so any future SDK/API change that flips one of them is caught here, not on a
// live cluster). If a method starts 404-ing, the client needs a not-found guard for it.
func TestAlicloudNasDescribeMissingResources(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	missingName := fmt.Sprintf("cm-does-not-exist-%d", time.Now().UnixNano())
	// NAS file system ids are short alphanumeric; use a plausible-but-absent one.
	missingFsId := fmt.Sprintf("00000000%d", time.Now().Unix()%100000000)

	t.Run("DescribeAccessGroups(missing)->empty,no-error", func(t *testing.T) {
		groups, err := client.DescribeAccessGroups(ctx, missingName)
		if err != nil {
			t.Fatalf("must not error for a missing access group, got: %v", err)
		}
		if len(groups) != 0 {
			t.Fatalf("expected empty, got: %+v", groups)
		}
	})

	t.Run("DescribeAccessRules(missing)->empty,no-error", func(t *testing.T) {
		rules, err := client.DescribeAccessRules(ctx, missingName)
		if err != nil {
			t.Fatalf("must not error for a missing access group, got: %v", err)
		}
		if len(rules) != 0 {
			t.Fatalf("expected empty, got: %+v", rules)
		}
	})

	t.Run("DescribeFileSystem(missing)->nil,no-error", func(t *testing.T) {
		fs, err := client.DescribeFileSystem(ctx, missingFsId)
		if err != nil {
			t.Fatalf("must not error for a missing file system, got: %v", err)
		}
		if fs != nil {
			t.Fatalf("expected nil for a missing file system, got: %+v", fs)
		}
	})

	t.Run("DescribeMountTargets(missing-fs)->empty,no-error", func(t *testing.T) {
		mts, err := client.DescribeMountTargets(ctx, missingFsId)
		if err != nil {
			t.Fatalf("must not error for a missing file system, got: %v", err)
		}
		if len(mts) != 0 {
			t.Fatalf("expected empty, got: %+v", mts)
		}
	})
}

func TestAlicloudNasMountTargetLifecycle(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	vpcId, vSwitchId, _, cleanup := createTempVpcWithVSwitch(t)
	defer cleanup()

	// Access group + rule
	agName := fmt.Sprintf("cm-nfs-test-%d", time.Now().UnixMilli())
	t.Logf("Creating access group %s ...", agName)
	if err := client.CreateAccessGroup(ctx, agName, "cloud-manager-integration-test"); err != nil {
		t.Fatalf("CreateAccessGroup failed: %v", err)
	}
	defer func() {
		if err := client.DeleteAccessGroup(ctx, agName); err != nil {
			t.Logf("Warning: DeleteAccessGroup failed: %v", err)
		}
	}()

	if err := client.CreateAccessRule(ctx, agName, "10.0.0.0/8"); err != nil {
		t.Fatalf("CreateAccessRule failed: %v", err)
	}
	rules, err := client.DescribeAccessRules(ctx, agName)
	if err != nil {
		t.Fatalf("DescribeAccessRules failed: %v", err)
	}
	t.Logf("Access rules: %v", rules)

	// File system
	fsId, err := client.CreateFileSystem(ctx, "NFS", "Performance", "")
	if err != nil {
		t.Fatalf("CreateFileSystem failed: %v", err)
	}
	defer func() {
		if err := client.DeleteFileSystem(ctx, fsId); err != nil {
			t.Logf("Warning: DeleteFileSystem failed: %v", err)
		}
	}()
	pollFileSystemStatus(t, client, ctx, fsId, "Running", 2*time.Minute)

	// Mount target
	t.Logf("Creating mount target fs=%s vpc=%s vsw=%s ...", fsId, vpcId, vSwitchId)
	domain, err := client.CreateMountTarget(ctx, fsId, vpcId, vSwitchId, agName)
	if err != nil {
		t.Fatalf("CreateMountTarget failed: %v", err)
	}
	t.Logf("Mount target created: %s", domain)

	mts, err := client.DescribeMountTargets(ctx, fsId)
	if err != nil {
		t.Fatalf("DescribeMountTargets failed: %v", err)
	}
	if len(mts) == 0 {
		t.Fatal("expected at least one mount target")
	}
	t.Logf("Mount targets: %+v", mts)

	// A mount target cannot be deleted while it is still creating
	// (VolumeStatusForbidOperation), so wait until it is Active.
	pollMountTargetStatus(t, client, ctx, fsId, domain, "Active", 2*time.Minute)

	// Delete mount target
	t.Logf("Deleting mount target %s ...", domain)
	if err := client.DeleteMountTarget(ctx, fsId, domain); err != nil {
		t.Fatalf("DeleteMountTarget failed: %v", err)
	}
	t.Log("Mount target deleted")

	// Wait until the mount target is fully gone before the deferred file system
	// deletion runs, otherwise it fails with MountTargetNotEmpty.
	pollMountTargetGone(t, client, ctx, fsId, domain, 2*time.Minute)
}

// pollMountTargetStatus waits until the given mount target reaches the desired status.
func pollMountTargetStatus(t *testing.T, client Client, ctx context.Context, fsId, domain, want string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		mts, err := client.DescribeMountTargets(ctx, fsId)
		if err != nil {
			t.Fatalf("DescribeMountTargets failed: %v", err)
		}
		for _, mt := range mts {
			if mt.MountTargetDomain == domain && mt.Status == want {
				return
			}
		}
		time.Sleep(3 * time.Second)
	}
	t.Fatalf("mount target %s did not reach status %q within %s", domain, want, timeout)
}

// pollMountTargetGone waits until the given mount target no longer exists.
func pollMountTargetGone(t *testing.T, client Client, ctx context.Context, fsId, domain string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		mts, err := client.DescribeMountTargets(ctx, fsId)
		if err != nil {
			t.Fatalf("DescribeMountTargets failed: %v", err)
		}
		found := false
		for _, mt := range mts {
			if mt.MountTargetDomain == domain {
				found = true
				break
			}
		}
		if !found {
			return
		}
		time.Sleep(3 * time.Second)
	}
	t.Fatalf("mount target %s still present after %s", domain, timeout)
}
