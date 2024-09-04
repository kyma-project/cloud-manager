package mock

import (
	"context"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elliotchance/pie/v2"
	"k8s.io/utils/ptr"
	"sync"
)

type RouteTableConfig interface {
	AddRouteTable(routeTableId, vpcId *string, tags []ec2types.Tag, associations []ec2types.RouteTableAssociation) ec2types.RouteTable
}
type routeTableEntry struct {
	routeTable ec2types.RouteTable
}
type routeTablesStore struct {
	m     sync.Mutex
	items []*routeTableEntry
}

func (s *routeTablesStore) AddRouteTable(routeTableId, vpcId *string, tags []ec2types.Tag, associations []ec2types.RouteTableAssociation) ec2types.RouteTable {
	s.m.Lock()
	defer s.m.Unlock()

	filtered := pie.Filter(s.items, func(entry *routeTableEntry) bool {
		return *entry.routeTable.RouteTableId == *routeTableId
	})

	entry := pie.First(filtered)

	if entry == nil {

		entry = &routeTableEntry{routeTable: ec2types.RouteTable{
			RouteTableId: routeTableId,
			Routes:       nil,
			Tags:         tags,
			VpcId:        vpcId,
			Associations: associations,
		}}

		s.items = append(s.items, entry)
	}

	return entry.routeTable
}
func (s *routeTablesStore) DescribeRouteTables(ctc context.Context, vpcId string) ([]ec2types.RouteTable, error) {
	s.m.Lock()
	defer s.m.Unlock()

	filtered := pie.Filter(s.items, func(e *routeTableEntry) bool {
		return *e.routeTable.VpcId == vpcId
	})

	return pie.Map(filtered, func(e *routeTableEntry) ec2types.RouteTable { return e.routeTable }), nil
}
func (s *routeTablesStore) CreateRoute(ctx context.Context, routeTableId, destinationCidrBlock, vpcPeeringConnectionId *string) error {
	s.m.Lock()
	defer s.m.Unlock()

	filtered := pie.Filter(s.items, func(r *routeTableEntry) bool {
		return *r.routeTable.RouteTableId == *routeTableId
	})

	entry := pie.First(filtered)

	entry.routeTable.Routes = append(entry.routeTable.Routes, ec2types.Route{
		DestinationCidrBlock:   destinationCidrBlock,
		VpcPeeringConnectionId: vpcPeeringConnectionId,
	})

	return nil
}

func (s *routeTablesStore) DeleteRoute(ctx context.Context, routeTableId, destinationCidrBlock *string) error {
	s.m.Lock()
	defer s.m.Unlock()

	filtered := pie.Filter(s.items, func(r *routeTableEntry) bool {
		return *r.routeTable.RouteTableId == *routeTableId
	})

	entry := pie.First(filtered)

	entry.routeTable.Routes = pie.Filter(entry.routeTable.Routes, func(r ec2types.Route) bool {
		return !ptr.Equal(r.DestinationCidrBlock, destinationCidrBlock)
	})

	return nil
}
