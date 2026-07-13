package mock

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/elliotchance/pie/v2"
	"github.com/google/uuid"
	alicloudiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/iprange/client"
	alicloudvpcnetworkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/vpcnetwork/client"
)

// VpcEntry is the stored representation of a VPC.
type VpcEntry struct {
	VpcId               string
	VpcName             string
	CidrBlock           string
	Status              string
	SecondaryCidrBlocks []string
}

// VSwitchEntry is the stored representation of a vSwitch.
type VSwitchEntry struct {
	VSwitchId   string
	VSwitchName string
	CidrBlock   string
	VpcId       string
	ZoneId      string
	Status      string
}

type vpcStore struct {
	m sync.Mutex

	vpcs      []*VpcEntry
	vSwitches []*VSwitchEntry
	zones     []string

	vpcErrors     map[string]error
	vSwitchErrors map[string]error
}

func newVpcStore() *vpcStore {
	return &vpcStore{
		vpcErrors:     map[string]error{},
		vSwitchErrors: map[string]error{},
	}
}

// === Config side (test seeding) =============================================

func (s *vpcStore) AddVpc(id, name, cidr string) *VpcEntry {
	s.m.Lock()
	defer s.m.Unlock()
	if id == "" {
		id = "vpc-" + uuid.NewString()[:8]
	}
	entry := &VpcEntry{VpcId: id, VpcName: name, CidrBlock: cidr, Status: "Available"}
	s.vpcs = append(s.vpcs, entry)
	return entry
}

func (s *vpcStore) AddVSwitch(vpcId, vSwitchId, name, zoneId, cidr string) *VSwitchEntry {
	s.m.Lock()
	defer s.m.Unlock()
	if vSwitchId == "" {
		vSwitchId = "vsw-" + uuid.NewString()[:8]
	}
	entry := &VSwitchEntry{VSwitchId: vSwitchId, VSwitchName: name, CidrBlock: cidr, VpcId: vpcId, ZoneId: zoneId, Status: "Available"}
	s.vSwitches = append(s.vSwitches, entry)
	return entry
}

func (s *vpcStore) AddZone(zoneId string) {
	s.m.Lock()
	defer s.m.Unlock()
	if slices.Contains(s.zones, zoneId) {
		return
	}
	s.zones = append(s.zones, zoneId)
}

func (s *vpcStore) SetVpcError(vpcId string, err error) {
	s.m.Lock()
	defer s.m.Unlock()
	if err == nil {
		delete(s.vpcErrors, vpcId)
	} else {
		s.vpcErrors[vpcId] = err
	}
}

func (s *vpcStore) SetVSwitchError(vSwitchId string, err error) {
	s.m.Lock()
	defer s.m.Unlock()
	if err == nil {
		delete(s.vSwitchErrors, vSwitchId)
	} else {
		s.vSwitchErrors[vSwitchId] = err
	}
}

// === Shared internal ========================================================

func (s *vpcStore) describeVpcsRaw(name string) []VpcEntry {
	s.m.Lock()
	defer s.m.Unlock()
	out := make([]VpcEntry, 0)
	for _, v := range s.vpcs {
		if name == "" || v.VpcName == name {
			out = append(out, *v)
		}
	}
	return out
}

// === vpcnetwork.Client: CreateVpc / DeleteVpc ================================

func (s *vpcStore) CreateVpc(ctx context.Context, name, cidrBlock string) (*alicloudvpcnetworkclient.VpcInfo, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	entry := s.AddVpc("", name, cidrBlock)
	return &alicloudvpcnetworkclient.VpcInfo{VpcId: entry.VpcId, VpcName: entry.VpcName, CidrBlock: entry.CidrBlock, Status: entry.Status}, nil
}

func (s *vpcStore) DeleteVpc(ctx context.Context, vpcId string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	if err, ok := s.vpcErrors[vpcId]; ok {
		return err
	}
	idx := pie.FindFirstUsing(s.vpcs, func(v *VpcEntry) bool { return v.VpcId == vpcId })
	if idx == -1 {
		return fmt.Errorf("vpc %s not found", vpcId)
	}
	for _, vsw := range s.vSwitches {
		if vsw.VpcId == vpcId {
			return errors.New("vpc has dependent vSwitches")
		}
	}
	s.vpcs = append(s.vpcs[:idx], s.vpcs[idx+1:]...)
	return nil
}

// === iprange.Client: vSwitch operations =====================================

func (s *vpcStore) CreateVSwitch(ctx context.Context, vpcId, zoneId, cidrBlock, name string) (string, error) {
	if isContextCanceled(ctx) {
		return "", context.Canceled
	}
	s.m.Lock()
	idx := pie.FindFirstUsing(s.vpcs, func(v *VpcEntry) bool { return v.VpcId == vpcId })
	if idx == -1 {
		s.m.Unlock()
		return "", fmt.Errorf("vpc %s not found", vpcId)
	}
	if err := s.checkVSwitchOverlap(vpcId, cidrBlock); err != nil {
		s.m.Unlock()
		return "", err
	}
	s.m.Unlock()
	entry := s.AddVSwitch(vpcId, "", name, zoneId, cidrBlock)
	return entry.VSwitchId, nil
}

func (s *vpcStore) checkVSwitchOverlap(vpcId, cidr string) error {
	for _, vsw := range s.vSwitches {
		if vsw.VpcId == vpcId && vsw.CidrBlock == cidr {
			return fmt.Errorf("specified CIDR block overlapped with other subnets. The CIDR was: %s", cidr)
		}
	}
	return nil
}

func (s *vpcStore) DescribeVSwitch(ctx context.Context, vSwitchId string) (*alicloudiprangeclient.VSwitchInfo, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	if err, ok := s.vSwitchErrors[vSwitchId]; ok {
		return nil, err
	}
	idx := pie.FindFirstUsing(s.vSwitches, func(v *VSwitchEntry) bool { return v.VSwitchId == vSwitchId })
	if idx == -1 {
		return nil, nil
	}
	v := s.vSwitches[idx]
	return &alicloudiprangeclient.VSwitchInfo{VSwitchId: v.VSwitchId, VSwitchName: v.VSwitchName, CidrBlock: v.CidrBlock, VpcId: v.VpcId, ZoneId: v.ZoneId, Status: v.Status}, nil
}

func (s *vpcStore) DescribeVSwitchesByName(ctx context.Context, vpcId, name string) ([]alicloudiprangeclient.VSwitchInfo, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	out := []alicloudiprangeclient.VSwitchInfo{}
	for _, v := range s.vSwitches {
		if v.VpcId == vpcId && v.VSwitchName == name {
			out = append(out, alicloudiprangeclient.VSwitchInfo{VSwitchId: v.VSwitchId, VSwitchName: v.VSwitchName, CidrBlock: v.CidrBlock, VpcId: v.VpcId, ZoneId: v.ZoneId, Status: v.Status})
		}
	}
	return out, nil
}

func (s *vpcStore) DescribeVSwitchesByVpcId(ctx context.Context, vpcId string) ([]alicloudiprangeclient.VSwitchInfo, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	var out []alicloudiprangeclient.VSwitchInfo
	for _, v := range s.vSwitches {
		if v.VpcId == vpcId {
			out = append(out, alicloudiprangeclient.VSwitchInfo{VSwitchId: v.VSwitchId, VSwitchName: v.VSwitchName, CidrBlock: v.CidrBlock, VpcId: v.VpcId, ZoneId: v.ZoneId, Status: v.Status})
		}
	}
	return out, nil
}

func (s *vpcStore) DeleteVSwitch(ctx context.Context, vSwitchId string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	if err, ok := s.vSwitchErrors[vSwitchId]; ok {
		return err
	}
	idx := pie.FindFirstUsing(s.vSwitches, func(v *VSwitchEntry) bool { return v.VSwitchId == vSwitchId })
	if idx == -1 {
		return fmt.Errorf("vswitch %s not found", vSwitchId)
	}
	s.vSwitches = append(s.vSwitches[:idx], s.vSwitches[idx+1:]...)
	return nil
}

func (s *vpcStore) DescribeZones(ctx context.Context) ([]string, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	out := make([]string, len(s.zones))
	copy(out, s.zones)
	return out, nil
}

func (s *vpcStore) DescribeVpcAttribute(ctx context.Context, vpcId string) (*alicloudiprangeclient.VpcAttributeInfo, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	idx := pie.FindFirstUsing(s.vpcs, func(v *VpcEntry) bool { return v.VpcId == vpcId })
	if idx == -1 {
		return nil, fmt.Errorf("vpc %s not found", vpcId)
	}
	v := s.vpcs[idx]
	out := make([]string, len(v.SecondaryCidrBlocks))
	copy(out, v.SecondaryCidrBlocks)
	return &alicloudiprangeclient.VpcAttributeInfo{VpcId: vpcId, SecondaryCidrBlocks: out}, nil
}

func (s *vpcStore) AssociateVpcCidrBlock(ctx context.Context, vpcId, cidrBlock string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	idx := pie.FindFirstUsing(s.vpcs, func(v *VpcEntry) bool { return v.VpcId == vpcId })
	if idx == -1 {
		return fmt.Errorf("vpc %s not found", vpcId)
	}
	if slices.Contains(s.vpcs[idx].SecondaryCidrBlocks, cidrBlock) {
		return nil
	}
	s.vpcs[idx].SecondaryCidrBlocks = append(s.vpcs[idx].SecondaryCidrBlocks, cidrBlock)
	return nil
}

func (s *vpcStore) UnassociateVpcCidrBlock(ctx context.Context, vpcId, cidrBlock string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	idx := pie.FindFirstUsing(s.vpcs, func(v *VpcEntry) bool { return v.VpcId == vpcId })
	if idx == -1 {
		return fmt.Errorf("vpc %s not found", vpcId)
	}
	s.vpcs[idx].SecondaryCidrBlocks = pie.Filter(s.vpcs[idx].SecondaryCidrBlocks, func(c string) bool { return c != cidrBlock })
	return nil
}
