//go:build integration

package client

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func newTestClient(t *testing.T) Client {
	t.Helper()
	accessKey := os.Getenv("ALICLOUD_ACCESS_KEY")
	secretKey := os.Getenv("ALICLOUD_SECRET_KEY")

	if accessKey == "" || secretKey == "" {
		t.Skip("ALICLOUD_ACCESS_KEY and ALICLOUD_SECRET_KEY must be set")
	}

	provider := NewClientProvider()
	client, err := provider(context.Background(), "eu-central-1", accessKey, secretKey)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	return client
}

func TestAlicloudVpcClientDescribe(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	vpcs, err := client.DescribeVpcs(ctx, "")
	if err != nil {
		t.Fatalf("DescribeVpcs failed (credentials may be invalid or lack permissions): %v", err)
	}

	t.Logf("SUCCESS: Connected to AliCloud. Found %d VPCs in eu-central-1", len(vpcs))
	for _, v := range vpcs {
		t.Logf("  VPC: id=%s name=%s cidr=%s status=%s", v.VpcId, v.VpcName, v.CidrBlock, v.Status)
	}
}

func TestAlicloudVpcClientCreateAndDelete(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	vpcName := fmt.Sprintf("cm-inttest-%d", time.Now().UnixMilli())
	cidr := "10.99.0.0/16"

	// Create
	t.Logf("Creating VPC name=%s cidr=%s ...", vpcName, cidr)
	vpcInfo, err := client.CreateVpc(ctx, vpcName, cidr)
	if err != nil {
		t.Fatalf("CreateVpc failed: %v", err)
	}
	t.Logf("VPC created: id=%s name=%s cidr=%s status=%s", vpcInfo.VpcId, vpcInfo.VpcName, vpcInfo.CidrBlock, vpcInfo.Status)

	// Wait a moment for the VPC to become available
	time.Sleep(3 * time.Second)

	// Describe - verify it exists
	vpcs, err := client.DescribeVpcs(ctx, vpcName)
	if err != nil {
		t.Fatalf("DescribeVpcs failed: %v", err)
	}
	if len(vpcs) == 0 {
		t.Fatalf("Expected to find VPC %s but got 0 results", vpcName)
	}
	t.Logf("Describe found VPC: id=%s status=%s", vpcs[0].VpcId, vpcs[0].Status)

	// Delete
	t.Logf("Deleting VPC id=%s ...", vpcInfo.VpcId)
	err = client.DeleteVpc(ctx, vpcInfo.VpcId)
	if err != nil {
		t.Fatalf("DeleteVpc failed: %v", err)
	}
	t.Logf("VPC deleted successfully")

	// Verify deletion
	time.Sleep(2 * time.Second)
	vpcs, err = client.DescribeVpcs(ctx, vpcName)
	if err != nil {
		t.Fatalf("DescribeVpcs after delete failed: %v", err)
	}
	t.Logf("After deletion: %d VPCs found with name %s", len(vpcs), vpcName)
}
