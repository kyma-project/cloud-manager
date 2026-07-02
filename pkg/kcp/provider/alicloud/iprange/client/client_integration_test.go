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

// createTempVpc creates a VPC for testing and returns its ID. Caller must delete it.
func createTempVpc(t *testing.T) string {
	t.Helper()
	config := &openapi.Config{
		AccessKeyId:     tea.String(os.Getenv("ALICLOUD_ACCESS_KEY")),
		AccessKeySecret: tea.String(os.Getenv("ALICLOUD_SECRET_KEY")),
		RegionId:        tea.String("eu-central-1"),
	}
	config.Endpoint = tea.String("vpc.eu-central-1.aliyuncs.com")
	vpcClient, err := vpc.NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create raw vpc client: %v", err)
	}
	name := fmt.Sprintf("cm-iprange-test-%d", time.Now().UnixMilli())
	resp, err := vpcClient.CreateVpc(&vpc.CreateVpcRequest{
		RegionId:  tea.String("eu-central-1"),
		VpcName:   tea.String(name),
		CidrBlock: tea.String("10.0.0.0/8"),
	})
	if err != nil {
		t.Fatalf("Failed to create temp VPC: %v", err)
	}
	vpcId := tea.StringValue(resp.Body.VpcId)
	t.Logf("Created temp VPC: %s (%s)", vpcId, name)
	time.Sleep(3 * time.Second) // wait for VPC to become Available
	return vpcId
}

// deleteTempVpc deletes a VPC created for testing.
func deleteTempVpc(t *testing.T, vpcId string) {
	t.Helper()
	config := &openapi.Config{
		AccessKeyId:     tea.String(os.Getenv("ALICLOUD_ACCESS_KEY")),
		AccessKeySecret: tea.String(os.Getenv("ALICLOUD_SECRET_KEY")),
		RegionId:        tea.String("eu-central-1"),
	}
	config.Endpoint = tea.String("vpc.eu-central-1.aliyuncs.com")
	vpcClient, err := vpc.NewClient(config)
	if err != nil {
		t.Logf("Warning: failed to create vpc client for cleanup: %v", err)
		return
	}
	_, err = vpcClient.DeleteVpc(&vpc.DeleteVpcRequest{
		RegionId: tea.String("eu-central-1"),
		VpcId:    tea.String(vpcId),
	})
	if err != nil {
		t.Logf("Warning: failed to delete temp VPC %s: %v", vpcId, err)
	} else {
		t.Logf("Deleted temp VPC: %s", vpcId)
	}
}

func TestAlicloudDescribeZones(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	zones, err := client.DescribeZones(ctx)
	if err != nil {
		t.Fatalf("DescribeZones failed: %v", err)
	}

	t.Logf("SUCCESS: Found %d zones in eu-central-1", len(zones))
	for _, z := range zones {
		t.Logf("  Zone: %s", z)
	}

	if len(zones) == 0 {
		t.Fatal("Expected at least one zone")
	}
}

func TestAlicloudVSwitchCreateAndDelete(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Create temp VPC
	vpcId := createTempVpc(t)
	defer deleteTempVpc(t, vpcId)

	// Get available zones
	zones, err := client.DescribeZones(ctx)
	if err != nil {
		t.Fatalf("DescribeZones failed: %v", err)
	}
	if len(zones) == 0 {
		t.Fatal("No zones available")
	}
	zoneId := zones[0]

	// Create VSwitch
	vSwitchName := fmt.Sprintf("cm-iprange-test-%d", time.Now().UnixMilli())
	cidr := "10.99.0.0/22"
	t.Logf("Creating VSwitch name=%s cidr=%s zone=%s vpc=%s ...", vSwitchName, cidr, zoneId, vpcId)

	vSwitchId, err := client.CreateVSwitch(ctx, vpcId, zoneId, cidr, vSwitchName)
	if err != nil {
		t.Fatalf("CreateVSwitch failed: %v", err)
	}
	t.Logf("VSwitch created: id=%s", vSwitchId)

	// Wait for VSwitch to become Available
	time.Sleep(2 * time.Second)

	// Describe by ID
	vsw, err := client.DescribeVSwitch(ctx, vSwitchId)
	if err != nil {
		t.Fatalf("DescribeVSwitch failed: %v", err)
	}
	if vsw == nil {
		t.Fatal("DescribeVSwitch returned nil")
	}
	t.Logf("Describe: id=%s name=%s cidr=%s zone=%s status=%s", vsw.VSwitchId, vsw.VSwitchName, vsw.CidrBlock, vsw.ZoneId, vsw.Status)

	// Describe by name
	vswitches, err := client.DescribeVSwitchesByName(ctx, vpcId, vSwitchName)
	if err != nil {
		t.Fatalf("DescribeVSwitchesByName failed: %v", err)
	}
	if len(vswitches) != 1 {
		t.Fatalf("Expected 1 VSwitch by name, got %d", len(vswitches))
	}

	// Delete
	t.Logf("Deleting VSwitch id=%s ...", vSwitchId)
	err = client.DeleteVSwitch(ctx, vSwitchId)
	if err != nil {
		t.Fatalf("DeleteVSwitch failed: %v", err)
	}
	t.Logf("VSwitch deleted successfully")

	// Verify deletion
	time.Sleep(2 * time.Second)
	vsw, err = client.DescribeVSwitch(ctx, vSwitchId)
	if err != nil {
		t.Logf("DescribeVSwitch after delete returned error (expected): %v", err)
	} else if vsw == nil {
		t.Logf("VSwitch confirmed deleted (nil)")
	} else {
		t.Logf("VSwitch after delete: status=%s", vsw.Status)
	}
}
