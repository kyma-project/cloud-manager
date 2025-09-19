package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/elliotchance/pie/v2"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/subnets"
	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/shareaccessrules"
	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/sharenetworks"
	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/shares"
	"github.com/gophercloud/gophercloud/v2/pagination"

	sapclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/client"
)

type Client interface {
	ListInternalNetworks(ctx context.Context, name string) ([]networks.Network, error)
	GetNetwork(ctx context.Context, id string) (*networks.Network, error)
	ListSubnets(ctx context.Context, networkId string) ([]subnets.Subnet, error)
	GetSubnet(ctx context.Context, id string) (*subnets.Subnet, error)

	ListShareNetworks(ctx context.Context, networkId string) ([]sharenetworks.ShareNetwork, error)
	GetShareNetwork(ctx context.Context, id string) (*sharenetworks.ShareNetwork, error)
	CreateShareNetwork(ctx context.Context, networkId, subnetId, name string) (*sharenetworks.ShareNetwork, error)
	DeleteShareNetwork(ctx context.Context, id string) error

	ListShares(ctx context.Context, shareNetworkId string) ([]Share, error)
	GetShare(ctx context.Context, id string) (*Share, error)
	CreateShare(ctx context.Context, shareNetworkId, name string, size int, snapshotID string, metadata map[string]string) (*Share, error)
	DeleteShare(ctx context.Context, id string) error
	ShareShrink(ctx context.Context, shareId string, newSize int) error
	ShareExtend(ctx context.Context, shareId string, newSize int) error

	ListShareAccessRules(ctx context.Context, shareId string) ([]ShareAccess, error)
	GrantShareAccess(ctx context.Context, shareId string, cidr string) (*ShareAccess, error)
	RevokeShareAccess(ctx context.Context, shareId, accessId string) error
}

var _ Client = &client{}

type client struct {
	//svc *gophercloud.ServiceClient
	netSvc   *gophercloud.ServiceClient
	shareSvc *gophercloud.ServiceClient
}

func NewClientProvider() sapclient.SapClientProvider[Client] {
	return func(ctx context.Context, pp sapclient.ProviderParams) (Client, error) {
		pi, err := sapclient.NewProviderClient(ctx, pp)
		if err != nil {
			return nil, fmt.Errorf("failed to create new sap provider client: %v", err)
		}
		netSvc, err := openstack.NewNetworkV2(pi.ProviderClient, pi.EndpointOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to create network v2 client: %v", err)
		}
		shareSvc, err := openstack.NewSharedFileSystemV2(pi.ProviderClient, pi.EndpointOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to create shared file system v2 client: %v", err)
		}
		// max microversion in xena is 2.65
		// https://docs.openstack.org/manila/latest/contributor/index.html
		// https://documentation.global.cloud.sap/docs/customer/support/faq-current-versions/
		shareSvc.Microversion = "2.65"
		return &client{
			netSvc:   netSvc,
			shareSvc: shareSvc,
		}, nil
	}
}

func (c *client) ListInternalNetworks(ctx context.Context, name string) ([]networks.Network, error) {
	pg, err := networks.List(c.netSvc, networks.ListOpts{
		Name: name,
	}).AllPages(ctx)
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error listing private networks: %w", err)
	}
	arr, err := networks.ExtractNetworks(pg)
	if err != nil {
		return nil, fmt.Errorf("error extracting private networks: %w", err)
	}
	return arr, nil
}

func (c *client) GetNetwork(ctx context.Context, id string) (*networks.Network, error) {
	n, err := networks.Get(ctx, c.netSvc, id).Extract()
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return nil, nil
	}
	return n, err
}

func (c *client) ListSubnets(ctx context.Context, networkId string) ([]subnets.Subnet, error) {
	pg, err := subnets.List(c.netSvc, subnets.ListOpts{
		NetworkID: networkId,
	}).AllPages(ctx)
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error listing subnets: %w", err)
	}
	arr, err := subnets.ExtractSubnets(pg)
	if err != nil {
		return nil, fmt.Errorf("error extracting subnets: %w", err)
	}

	return arr, nil
}

func (c *client) GetSubnet(ctx context.Context, id string) (*subnets.Subnet, error) {
	subnet, err := subnets.Get(ctx, c.netSvc, id).Extract()
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting subnet: %w", err)
	}

	return subnet, nil
}

// Share Networks ------------------------------------------------------------------------------

func (c *client) ListShareNetworks(ctx context.Context, networkId string) ([]sharenetworks.ShareNetwork, error) {
	pg, err := sharenetworks.ListDetail(c.shareSvc, sharenetworks.ListOpts{
		NeutronNetID: networkId,
	}).AllPages(ctx)
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error listing sharenetworks: %w", err)
	}
	arr, err := sharenetworks.ExtractShareNetworks(pg)
	if err != nil {
		return nil, fmt.Errorf("error extracting sharenetworks: %w", err)
	}

	return arr, nil
}

func (c *client) GetShareNetwork(ctx context.Context, id string) (*sharenetworks.ShareNetwork, error) {
	net, err := sharenetworks.Get(ctx, c.shareSvc, id).Extract()
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting sharenetwork: %w", err)
	}
	return net, nil
}

func (c *client) CreateShareNetwork(ctx context.Context, networkId, subnetId, name string) (*sharenetworks.ShareNetwork, error) {
	net, err := sharenetworks.Create(ctx, c.shareSvc, sharenetworks.CreateOpts{
		NeutronNetID:    networkId,
		NeutronSubnetID: subnetId,
		Name:            name,
	}).Extract()
	if err != nil {
		return net, fmt.Errorf("error creating sharenetwork: %w", err)
	}
	return net, nil
}

func (c *client) DeleteShareNetwork(ctx context.Context, id string) error {
	err := sharenetworks.Delete(ctx, c.shareSvc, id).ExtractErr()
	if err != nil {
		return fmt.Errorf("error deleting sharenetwork: %w", err)
	}
	return nil
}

// shares ----------------------------------------------------------------------------------

type Share struct {
	// The availability zone of the share
	AvailabilityZone string `json:"availability_zone"`
	// A description of the share
	Description string `json:"description,omitempty"`
	// DisplayDescription is inherited from BlockStorage API.
	// Both Description and DisplayDescription can be used
	DisplayDescription string `json:"display_description,omitempty"`
	// DisplayName is inherited from BlockStorage API
	// Both DisplayName and Name can be used
	DisplayName string `json:"display_name,omitempty"`
	// Indicates whether a share has replicas or not.
	HasReplicas bool `json:"has_replicas"`
	// The host name of the share
	Host string `json:"host"`
	// The UUID of the share
	ID string `json:"id"`
	// Indicates the visibility of the share
	IsPublic bool `json:"is_public,omitempty"`
	// Share links for pagination
	Links []map[string]string `json:"links"`
	// Key, value -pairs of custom metadata
	Metadata map[string]string `json:"metadata,omitempty"`
	// The name of the share
	Name string `json:"name,omitempty"`
	// The UUID of the project to which this share belongs to
	ProjectID string `json:"project_id"`
	// The share replication type
	ReplicationType string `json:"replication_type,omitempty"`
	// The UUID of the share network
	ShareNetworkID string `json:"share_network_id"`
	// The shared file system protocol
	ShareProto string `json:"share_proto"`
	// The UUID of the share server
	ShareServerID string `json:"share_server_id"`
	// The UUID of the share type.
	ShareType string `json:"share_type"`
	// The name of the share type.
	ShareTypeName string `json:"share_type_name"`
	// The UUID of the share group. Available starting from the microversion 2.31
	ShareGroupID string `json:"share_group_id"`
	// Size of the share in GB
	Size int `json:"size"`
	// UUID of the snapshot from which to create the share
	SnapshotID string `json:"snapshot_id"`
	// The share status
	Status string `json:"status"`
	// The task state, used for share migration
	TaskState string `json:"task_state"`
	// The type of the volume
	VolumeType string `json:"volume_type,omitempty"`
	// The UUID of the consistency group this share belongs to
	ConsistencyGroupID string `json:"consistency_group_id"`
	// Used for filtering backends which either support or do not support share snapshots
	SnapshotSupport          bool   `json:"snapshot_support"`
	SourceCgsnapshotMemberID string `json:"source_cgsnapshot_member_id"`
	// Used for filtering backends which either support or do not support creating shares from snapshots
	CreateShareFromSnapshotSupport bool `json:"create_share_from_snapshot_support"`
	// Timestamp when the share was created
	CreatedAt time.Time `json:"-"`
	// Timestamp when the share was updated
	UpdatedAt time.Time `json:"-"`

	ExportLocation  string   `json:"export_location"`
	ExportLocations []string `json:"export_locations"`
}

// share.status possible values https://docs.openstack.org/manila/latest/user/create-and-manage-shares.html
// These “-ing” states end in a “available” state if everything goes well. They may end up in an “error” state in case there is an issue.
// * available
// * error
// * creating
// * extending
// * shrinking
// * migrating

func (c *client) ListShares(ctx context.Context, shareNetworkId string) ([]Share, error) {
	pg, err := shares.ListDetail(c.shareSvc, shares.ListOpts{
		ShareNetworkID: shareNetworkId,
	}).AllPages(ctx)
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error listing shares: %w", err)
	}
	arr, err := extractShares(pg)
	if err != nil {
		return nil, fmt.Errorf("error extracting shares: %w", err)
	}
	return arr, nil
}

func extractShares(r pagination.Page) ([]Share, error) {
	var s struct {
		Shares []Share `json:"shares"`
	}

	err := (r.(shares.SharePage)).ExtractInto(&s)

	return s.Shares, err
}

func newShareFromGopherShare(s *shares.Share) *Share {
	return &Share{
		AvailabilityZone:               s.AvailabilityZone,
		Description:                    s.Description,
		DisplayDescription:             s.DisplayDescription,
		DisplayName:                    s.DisplayName,
		HasReplicas:                    s.HasReplicas,
		Host:                           s.Host,
		ID:                             s.ID,
		IsPublic:                       s.IsPublic,
		Links:                          s.Links,
		Metadata:                       s.Metadata,
		Name:                           s.Name,
		ProjectID:                      s.ProjectID,
		ReplicationType:                s.ReplicationType,
		ShareNetworkID:                 s.ShareNetworkID,
		ShareProto:                     s.ShareProto,
		ShareServerID:                  s.ShareServerID,
		ShareType:                      s.ShareType,
		ShareTypeName:                  s.ShareType,
		ShareGroupID:                   s.ShareGroupID,
		Size:                           s.Size,
		SnapshotID:                     s.SnapshotID,
		Status:                         s.Status,
		TaskState:                      s.TaskState,
		VolumeType:                     s.VolumeType,
		ConsistencyGroupID:             s.ConsistencyGroupID,
		SnapshotSupport:                s.SnapshotSupport,
		SourceCgsnapshotMemberID:       s.SourceCgsnapshotMemberID,
		CreateShareFromSnapshotSupport: s.CreateShareFromSnapshotSupport,
		CreatedAt:                      s.CreatedAt,
		UpdatedAt:                      s.UpdatedAt,
		ExportLocation:                 "",
		ExportLocations:                nil,
	}
}
func (c *client) GetShare(ctx context.Context, id string) (*Share, error) {
	var s struct {
		Share *Share `json:"share"`
	}
	//sh := &Share{}
	err := shares.Get(ctx, c.shareSvc, id).ExtractInto(&s)
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting share: %w", err)
	}
	return s.Share, nil
}

func (c *client) CreateShare(ctx context.Context, shareNetworkId, name string, size int, snapshotID string, metadata map[string]string) (*Share, error) {
	sh, err := shares.Create(ctx, c.shareSvc, shares.CreateOpts{
		ShareProto:     "NFS",
		Size:           size,
		Name:           name,
		ShareNetworkID: shareNetworkId,
		SnapshotID:     snapshotID,
		Metadata:       metadata,
	}).Extract()
	if err != nil {
		return nil, fmt.Errorf("error creating share: %w", err)
	}
	return newShareFromGopherShare(sh), nil
}

func (c *client) DeleteShare(ctx context.Context, id string) error {
	err := shares.Delete(ctx, c.shareSvc, id).ExtractErr()
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting share: %w", err)
	}
	return nil
}

type ExportLocation struct {
	ShareInstanceID string
	Path            string
	Preferred       bool
	ID              string
}

func (c *client) ShareShrink(ctx context.Context, shareId string, newSize int) error {
	err := shares.Shrink(ctx, c.shareSvc, shareId, shares.ShrinkOpts{
		NewSize: newSize,
	}).ExtractErr()
	if err != nil {
		return err
	}
	return nil
}

func (c *client) ShareExtend(ctx context.Context, shareId string, newSize int) error {
	err := shares.Extend(ctx, c.shareSvc, shareId, shares.ExtendOpts{
		NewSize: newSize,
	}).ExtractErr()
	if err != nil {
		return err
	}
	return nil
}

// share access -------------------------------------------------------------------

type ShareAccess struct {
	ID          string
	ShareID     string
	AccessType  string
	AccessTo    string
	AccessKey   string
	State       string
	AccessLevel string
}

func newShareAccessFromSharesAccessRight(o *shares.AccessRight) *ShareAccess {
	return &ShareAccess{
		ID:          o.ID,
		ShareID:     o.ShareID,
		AccessType:  o.AccessType,
		AccessTo:    o.AccessTo,
		AccessKey:   o.AccessKey,
		State:       o.State,
		AccessLevel: o.AccessLevel,
	}
}

func newShareAccessFromShareAccessRulesShareAccess(o *shareaccessrules.ShareAccess) *ShareAccess {
	return &ShareAccess{
		ID:          o.ID,
		ShareID:     o.ShareID,
		AccessType:  o.AccessType,
		AccessTo:    o.AccessTo,
		AccessKey:   o.AccessKey,
		State:       o.State,
		AccessLevel: o.AccessLevel,
	}
}

func (c *client) ListShareAccessRules(ctx context.Context, shareId string) ([]ShareAccess, error) {
	arr, err := shareaccessrules.List(ctx, c.shareSvc, shareId).Extract()
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error listing access rights: %w", err)
	}
	return pie.Map(arr, func(x shareaccessrules.ShareAccess) ShareAccess {
		x.ShareID = shareId
		return *newShareAccessFromShareAccessRulesShareAccess(&x)
	}), nil
}

// GrantAccessOpts is a temporary fix until CCloud upgrades above xena manilla release
// when this is called `allow_access` and is supported in gophercloud as microversion 2.7
// Once CCloud upgrades to yoga we can start using microversion 2.7 and regular gophercloud
// argument type `shares.GrantAccessOpts`
// https://documentation.global.cloud.sap/docs/customer/support/faq-current-versions/
// https://docs.openstack.org/api-ref/shared-file-system/#grant-access
// https://docs.openstack.org/manila/latest/contributor/index.html
// https://github.com/gophercloud/gophercloud/issues/488
// https://github.com/gophercloud/gophercloud/issues/441
type GrantAccessOpts shares.GrantAccessOpts

// ToGrantAccessMap overrides the shares.GrantAccessOpts method so it can rename the parent to `os-allow_access`
// that is in microversions <2.7 as currently used by CCloud in the xena release
func (opts GrantAccessOpts) ToGrantAccessMap() (map[string]any, error) {
	return gophercloud.BuildRequestBody(opts, "os-allow_access")
}

func (c *client) GrantShareAccess(ctx context.Context, shareId string, cidr string) (*ShareAccess, error) {
	ar, err := shares.GrantAccess(ctx, c.shareSvc, shareId, GrantAccessOpts{
		AccessType:  "ip",
		AccessTo:    cidr,
		AccessLevel: "rw",
	}).Extract()
	if err != nil {
		return nil, fmt.Errorf("error granting access to share: %w", err)
	}
	ar.ShareID = shareId
	return newShareAccessFromSharesAccessRight(ar), nil
}

func (c *client) RevokeShareAccess(ctx context.Context, shareId, accessId string) error {
	err := shares.RevokeAccess(ctx, c.shareSvc, shareId, shares.RevokeAccessOpts{
		AccessID: accessId,
	}).ExtractErr()
	if err != nil {
		return fmt.Errorf("error revoking access to share: %w", err)
	}
	return nil
}
