package cloud

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"sort"
	"strings"
	"time"

	ecs "github.com/alibabacloud-go/ecs-20140526/v7/client"
	"github.com/alibabacloud-go/tea/dara"
	vpc "github.com/alibabacloud-go/vpc-20160428/v6/client"

	"cephtower/tools/aliyun-ceph-lab/internal/config"
	"cephtower/tools/aliyun-ceph-lab/internal/logging"
	"cephtower/tools/aliyun-ceph-lab/internal/state"
)

const (
	managedByTagKey         = "ceph-lab-managed-by"
	managedByTagValue       = "cephtower"
	clusterTagKey           = "ceph-lab-cluster"
	vpcNamePrefix           = "ceph-vpc-"
	vSwitchNamePrefix       = "ceph-switch-"
	securityGroupNamePrefix = "ceph-security-group-"
)

type vpcInfo struct {
	ID     string
	Name   string
	Status string
	CIDR   string
}

type vSwitchInfo struct {
	ID          string
	Name        string
	VPCID       string
	ZoneID      string
	Status      string
	AvailableIP int64
}

type securityGroupInfo struct {
	ID             string
	Name           string
	VPCID          string
	ServiceManaged bool
}

func (c *Client) EnsureNetwork(
	ctx context.Context,
	cfg *config.Config,
	tokenTime time.Time,
	onProgress func(state.Network) error,
) (state.Network, error) {
	var result state.Network

	var selectedVSwitch *vSwitchInfo
	if cfg.VSwitchID != "" {
		item, err := c.describeVSwitch(ctx, cfg.RegionID, cfg.VSwitchID)
		if err != nil {
			return result, err
		}
		if item.Status == "Pending" {
			if err := c.waitVSwitchAvailable(ctx, cfg.RegionID, cfg.ZoneID, item.ID); err != nil {
				return result, err
			}
			item, err = c.describeVSwitch(ctx, cfg.RegionID, cfg.VSwitchID)
			if err != nil {
				return result, err
			}
		}
		if err := validateVSwitch(item, cfg.ZoneID, int64(len(cfg.Nodes))); err != nil {
			return result, err
		}
		selectedVSwitch = &item
		result.VSwitchID = item.ID
		result.VPCID = item.VPCID
	}

	var selectedSecurityGroup *securityGroupInfo
	if cfg.SecurityGroupID != "" {
		item, err := c.describeSecurityGroup(ctx, cfg.RegionID, cfg.SecurityGroupID)
		if err != nil {
			return result, err
		}
		selectedSecurityGroup = &item
		result.SecurityGroupID = item.ID
		if result.VPCID != "" && result.VPCID != item.VPCID {
			return result, fmt.Errorf("security group %s and vSwitch %s belong to different VPCs", item.ID, result.VSwitchID)
		}
		result.VPCID = item.VPCID
	}

	if cfg.VPCID != "" {
		if result.VPCID != "" && result.VPCID != cfg.VPCID {
			return result, fmt.Errorf("configured vpc_id %s does not match the VPC inferred from network resources (%s)", cfg.VPCID, result.VPCID)
		}
		item, err := c.describeVPC(ctx, cfg.RegionID, cfg.VPCID)
		if err != nil {
			return result, err
		}
		if item.Status == "Pending" {
			if err := c.waitVPCAvailable(ctx, cfg.RegionID, item.ID); err != nil {
				return result, err
			}
		} else if item.Status != "Available" {
			return result, fmt.Errorf("VPC %s is not Available (status %s)", item.ID, item.Status)
		}
		result.VPCID = item.ID
	}

	if result.VPCID == "" && cfg.NetworkReuseManagedResources() {
		item, found, err := c.findManagedVPC(ctx, cfg)
		if err != nil {
			return result, err
		}
		if found {
			logging.Infof("network: reusing managed VPC %s", item.ID)
			result.VPCID = item.ID
			if item.Status != "Available" {
				if err := c.waitVPCAvailable(ctx, cfg.RegionID, item.ID); err != nil {
					return result, err
				}
			}
		}
	}
	if result.VPCID == "" {
		if !cfg.NetworkAutoCreate() {
			return result, errors.New("no managed VPC was found and network.auto_create is false")
		}
		id, err := c.createVPC(ctx, cfg, tokenTime)
		if err != nil {
			return result, err
		}
		result.VPCID, result.CreatedVPC = id, true
		if err := onProgress(result); err != nil {
			return result, err
		}
		logging.Infof("network: created VPC %s", id)
		if err := c.waitVPCAvailable(ctx, cfg.RegionID, id); err != nil {
			return result, err
		}
	}

	if selectedVSwitch == nil && cfg.NetworkReuseManagedResources() {
		item, found, err := c.findManagedVSwitch(ctx, cfg, result.VPCID)
		if err != nil {
			return result, err
		}
		if found {
			logging.Infof("network: reusing managed vSwitch %s", item.ID)
			if item.Status != "Available" {
				if err := c.waitVSwitchAvailable(ctx, cfg.RegionID, cfg.ZoneID, item.ID); err != nil {
					return result, err
				}
				item, err = c.describeVSwitch(ctx, cfg.RegionID, item.ID)
				if err != nil {
					return result, err
				}
			}
			if err := validateVSwitch(item, cfg.ZoneID, int64(len(cfg.Nodes))); err != nil {
				return result, err
			}
			selectedVSwitch = &item
			result.VSwitchID = item.ID
		}
	}
	if selectedVSwitch == nil {
		if !cfg.NetworkAutoCreate() {
			return result, errors.New("no suitable managed vSwitch was found and network.auto_create is false")
		}
		id, err := c.createVSwitch(ctx, cfg, result.VPCID, tokenTime)
		if err != nil {
			return result, err
		}
		result.VSwitchID, result.CreatedVSwitch = id, true
		if err := onProgress(result); err != nil {
			return result, err
		}
		logging.Infof("network: created vSwitch %s", id)
		if err := c.waitVSwitchAvailable(ctx, cfg.RegionID, cfg.ZoneID, id); err != nil {
			return result, err
		}
		item, err := c.describeVSwitch(ctx, cfg.RegionID, id)
		if err != nil {
			return result, err
		}
		if err := validateVSwitch(item, cfg.ZoneID, int64(len(cfg.Nodes))); err != nil {
			return result, err
		}
	}

	if selectedSecurityGroup == nil && cfg.NetworkReuseManagedResources() {
		item, found, err := c.findManagedSecurityGroup(ctx, cfg, result.VPCID)
		if err != nil {
			return result, err
		}
		if found {
			logging.Infof("network: reusing managed security group %s", item.ID)
			selectedSecurityGroup = &item
			result.SecurityGroupID = item.ID
			if cfg.Network.AccessSourceCIDR != "" {
				if err := c.authorizeLabSecurityGroup(ctx, cfg, item.ID, tokenTime); err != nil {
					return result, err
				}
			}
		}
	}
	if selectedSecurityGroup == nil {
		if !cfg.NetworkAutoCreate() {
			return result, errors.New("no managed security group was found and network.auto_create is false")
		}
		if cfg.Network.AccessSourceCIDR == "" {
			return result, errors.New("network.access_source_cidr is required to create a security group")
		}
		id, err := c.createSecurityGroup(ctx, cfg, result.VPCID, tokenTime)
		if err != nil {
			return result, err
		}
		result.SecurityGroupID, result.CreatedSecurityGroup = id, true
		if err := onProgress(result); err != nil {
			return result, err
		}
		logging.Infof("network: created security group %s", id)
		if err := c.authorizeLabSecurityGroup(ctx, cfg, id, tokenTime); err != nil {
			return result, err
		}
	}

	if err := onProgress(result); err != nil {
		return result, err
	}
	return result, nil
}

func (c *Client) DeleteNetwork(
	ctx context.Context,
	regionID string,
	network state.Network,
	onProgress func(state.Network) error,
) error {
	if network.CreatedSecurityGroup && network.SecurityGroupID != "" {
		logging.Infof("network: deleting security group %s", network.SecurityGroupID)
		if _, err := withCloudRetry(ctx, "DeleteSecurityGroup "+network.SecurityGroupID, func() (*ecs.DeleteSecurityGroupResponse, error) {
			return c.ecs.DeleteSecurityGroupWithOptions(&ecs.DeleteSecurityGroupRequest{
				RegionId: dara.String(regionID), SecurityGroupId: dara.String(network.SecurityGroupID),
			}, c.runtime)
		}); err != nil {
			return fmt.Errorf("DeleteSecurityGroup %s: %w", network.SecurityGroupID, err)
		}
		network.CreatedSecurityGroup = false
		if err := onProgress(network); err != nil {
			return err
		}
	}
	if network.CreatedVSwitch && network.VSwitchID != "" {
		logging.Infof("network: deleting vSwitch %s", network.VSwitchID)
		if _, err := withCloudRetry(ctx, "DeleteVSwitch "+network.VSwitchID, func() (*vpc.DeleteVSwitchResponse, error) {
			return c.vpc.DeleteVSwitchWithOptions(&vpc.DeleteVSwitchRequest{
				RegionId: dara.String(regionID), VSwitchId: dara.String(network.VSwitchID),
			}, c.runtime)
		}); err != nil {
			return fmt.Errorf("DeleteVSwitch %s: %w", network.VSwitchID, err)
		}
		if err := c.waitVSwitchDeleted(ctx, regionID, network.VSwitchID); err != nil {
			return err
		}
		network.CreatedVSwitch = false
		if err := onProgress(network); err != nil {
			return err
		}
	}
	if network.CreatedVPC && network.VPCID != "" {
		logging.Infof("network: deleting VPC %s", network.VPCID)
		if _, err := withCloudRetry(ctx, "DeleteVpc "+network.VPCID, func() (*vpc.DeleteVpcResponse, error) {
			return c.vpc.DeleteVpcWithOptions(&vpc.DeleteVpcRequest{
				RegionId: dara.String(regionID), VpcId: dara.String(network.VPCID),
			}, c.runtime)
		}); err != nil {
			return fmt.Errorf("DeleteVpc %s: %w", network.VPCID, err)
		}
		network.CreatedVPC = false
		if err := onProgress(network); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) describeVPC(ctx context.Context, regionID, id string) (vpcInfo, error) {
	if err := ctx.Err(); err != nil {
		return vpcInfo{}, err
	}
	response, err := withCloudRetry(ctx, "DescribeVpcs "+id, func() (*vpc.DescribeVpcsResponse, error) {
		return c.vpc.DescribeVpcsWithOptions(&vpc.DescribeVpcsRequest{
			RegionId: dara.String(regionID), VpcId: dara.String(id), PageSize: dara.Int32(50),
		}, c.runtime)
	})
	if err != nil {
		return vpcInfo{}, fmt.Errorf("DescribeVpcs %s: %w", id, err)
	}
	items := vpcItems(response)
	if len(items) != 1 || items[0].VpcId == nil {
		return vpcInfo{}, fmt.Errorf("VPC %s was not found in region %s", id, regionID)
	}
	return vpcInfo{
		ID: *items[0].VpcId, Name: stringValue(items[0].VpcName),
		Status: stringValue(items[0].Status), CIDR: stringValue(items[0].CidrBlock),
	}, nil
}

func (c *Client) findManagedVPC(ctx context.Context, cfg *config.Config) (vpcInfo, bool, error) {
	var candidates []vpcInfo
	for page := int32(1); ; page++ {
		if err := ctx.Err(); err != nil {
			return vpcInfo{}, false, err
		}
		response, err := withCloudRetry(ctx, fmt.Sprintf("DescribeVpcs page %d", page), func() (*vpc.DescribeVpcsResponse, error) {
			return c.vpc.DescribeVpcsWithOptions(&vpc.DescribeVpcsRequest{
				RegionId: dara.String(cfg.RegionID), PageNumber: dara.Int32(page),
				PageSize: dara.Int32(50), IsDefault: dara.Bool(false),
			}, c.runtime)
		})
		if err != nil {
			return vpcInfo{}, false, fmt.Errorf("find VPC with prefix %s: %w", vpcNamePrefix, err)
		}
		items := vpcItems(response)
		for _, item := range items {
			if item != nil && item.VpcId != nil && strings.HasPrefix(stringValue(item.VpcName), vpcNamePrefix) &&
				stringValue(item.CidrBlock) == cfg.Network.VPCCIDR &&
				(stringValue(item.Status) == "Available" || stringValue(item.Status) == "Pending") {
				candidates = append(candidates, vpcInfo{
					ID: *item.VpcId, Name: stringValue(item.VpcName),
					Status: stringValue(item.Status), CIDR: stringValue(item.CidrBlock),
				})
			}
		}
		if len(items) < 50 || response == nil || response.Body == nil ||
			page*50 >= int32Value(response.Body.TotalCount) {
			break
		}
	}
	if len(candidates) == 0 {
		return vpcInfo{}, false, nil
	}
	preferredName := resourceName(vpcNamePrefix, cfg.ClusterName)
	sort.Slice(candidates, func(i, j int) bool {
		if (candidates[i].Name == preferredName) != (candidates[j].Name == preferredName) {
			return candidates[i].Name == preferredName
		}
		return candidates[i].ID < candidates[j].ID
	})
	return candidates[0], true, nil
}

func (c *Client) createVPC(ctx context.Context, cfg *config.Config, tokenTime time.Time) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	response, err := withCloudRetry(ctx, "CreateVpc", func() (*vpc.CreateVpcResponse, error) {
		return c.vpc.CreateVpcWithOptions(&vpc.CreateVpcRequest{
			RegionId:    dara.String(cfg.RegionID),
			CidrBlock:   dara.String(cfg.Network.VPCCIDR),
			VpcName:     dara.String(resourceName(vpcNamePrefix, cfg.ClusterName)),
			Description: dara.String("Ceph lab network managed by CephTower"),
			ClientToken: dara.String(clientToken(cfg.ClusterName, "vpc", tokenTime)),
			Tag: []*vpc.CreateVpcRequestTag{
				{Key: dara.String(managedByTagKey), Value: dara.String(managedByTagValue)},
				{Key: dara.String(clusterTagKey), Value: dara.String(cfg.ClusterName)},
			},
		}, c.runtime)
	})
	if err != nil {
		return "", fmt.Errorf("CreateVpc: %w", err)
	}
	if response == nil || response.Body == nil || response.Body.VpcId == nil {
		return "", errors.New("CreateVpc returned no VPC ID")
	}
	return *response.Body.VpcId, nil
}

func (c *Client) describeVSwitch(ctx context.Context, regionID, id string) (vSwitchInfo, error) {
	if err := ctx.Err(); err != nil {
		return vSwitchInfo{}, err
	}
	response, err := withCloudRetry(ctx, "DescribeVSwitches "+id, func() (*vpc.DescribeVSwitchesResponse, error) {
		return c.vpc.DescribeVSwitchesWithOptions(&vpc.DescribeVSwitchesRequest{
			RegionId: dara.String(regionID), VSwitchId: dara.String(id), PageSize: dara.Int32(50),
		}, c.runtime)
	})
	if err != nil {
		return vSwitchInfo{}, fmt.Errorf("DescribeVSwitches %s: %w", id, err)
	}
	items := vSwitchItems(response)
	if len(items) != 1 || items[0].VSwitchId == nil {
		return vSwitchInfo{}, fmt.Errorf("vSwitch %s was not found in region %s", id, regionID)
	}
	item := items[0]
	result := vSwitchInfo{
		ID: *item.VSwitchId, Name: stringValue(item.VSwitchName),
		VPCID: stringValue(item.VpcId), ZoneID: stringValue(item.ZoneId),
		Status: stringValue(item.Status), AvailableIP: int64Value(item.AvailableIpAddressCount),
	}
	if result.VPCID == "" {
		return vSwitchInfo{}, fmt.Errorf("vSwitch %s is not attached to a VPC", id)
	}
	return result, nil
}

func (c *Client) findManagedVSwitch(ctx context.Context, cfg *config.Config, vpcID string) (vSwitchInfo, bool, error) {
	var candidates []vSwitchInfo
	for page := int32(1); ; page++ {
		if err := ctx.Err(); err != nil {
			return vSwitchInfo{}, false, err
		}
		response, err := withCloudRetry(ctx, fmt.Sprintf("DescribeVSwitches page %d", page), func() (*vpc.DescribeVSwitchesResponse, error) {
			return c.vpc.DescribeVSwitchesWithOptions(&vpc.DescribeVSwitchesRequest{
				RegionId: dara.String(cfg.RegionID), VpcId: dara.String(vpcID), ZoneId: dara.String(cfg.ZoneID),
				PageNumber: dara.Int32(page), PageSize: dara.Int32(50),
			}, c.runtime)
		})
		if err != nil {
			return vSwitchInfo{}, false, fmt.Errorf("find vSwitch with prefix %s: %w", vSwitchNamePrefix, err)
		}
		items := vSwitchItems(response)
		for _, item := range items {
			if item == nil || item.VSwitchId == nil ||
				!strings.HasPrefix(stringValue(item.VSwitchName), vSwitchNamePrefix) {
				continue
			}
			candidate := vSwitchInfo{
				ID: *item.VSwitchId, Name: stringValue(item.VSwitchName),
				VPCID: stringValue(item.VpcId), ZoneID: stringValue(item.ZoneId),
				Status: stringValue(item.Status), AvailableIP: int64Value(item.AvailableIpAddressCount),
			}
			if validateVSwitch(candidate, cfg.ZoneID, int64(len(cfg.Nodes))) == nil ||
				(candidate.ZoneID == cfg.ZoneID && candidate.VPCID != "" && candidate.Status == "Pending") {
				candidates = append(candidates, candidate)
			}
		}
		if len(items) < 50 || response == nil || response.Body == nil ||
			page*50 >= int32Value(response.Body.TotalCount) {
			break
		}
	}
	if len(candidates) == 0 {
		return vSwitchInfo{}, false, nil
	}
	preferredName := resourceName(vSwitchNamePrefix, cfg.ClusterName)
	sort.Slice(candidates, func(i, j int) bool {
		if (candidates[i].Name == preferredName) != (candidates[j].Name == preferredName) {
			return candidates[i].Name == preferredName
		}
		return candidates[i].ID < candidates[j].ID
	})
	return candidates[0], true, nil
}

func (c *Client) createVSwitch(ctx context.Context, cfg *config.Config, vpcID string, tokenTime time.Time) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	vpcItem, err := c.describeVPC(ctx, cfg.RegionID, vpcID)
	if err != nil {
		return "", err
	}
	if vpcItem.Status == "Pending" {
		if err := c.waitVPCAvailable(ctx, cfg.RegionID, vpcID); err != nil {
			return "", err
		}
		vpcItem, err = c.describeVPC(ctx, cfg.RegionID, vpcID)
		if err != nil {
			return "", err
		}
	}
	if vpcItem.Status != "Available" {
		return "", fmt.Errorf("VPC %s is not Available (status %s)", vpcID, vpcItem.Status)
	}
	if !cidrContains(vpcItem.CIDR, cfg.Network.VSwitchCIDR) {
		return "", fmt.Errorf("configured vSwitch CIDR %s is not a subnet of VPC %s CIDR %s", cfg.Network.VSwitchCIDR, vpcID, vpcItem.CIDR)
	}
	response, err := withCloudRetry(ctx, "CreateVSwitch", func() (*vpc.CreateVSwitchResponse, error) {
		return c.vpc.CreateVSwitchWithOptions(&vpc.CreateVSwitchRequest{
			RegionId:    dara.String(cfg.RegionID),
			ZoneId:      dara.String(cfg.ZoneID),
			VpcId:       dara.String(vpcID),
			CidrBlock:   dara.String(cfg.Network.VSwitchCIDR),
			VSwitchName: dara.String(resourceName(vSwitchNamePrefix, cfg.ClusterName)),
			Description: dara.String("Ceph lab subnet managed by CephTower"),
			ClientToken: dara.String(clientToken(cfg.ClusterName, "vswitch", tokenTime)),
			Tag: []*vpc.CreateVSwitchRequestTag{
				{Key: dara.String(managedByTagKey), Value: dara.String(managedByTagValue)},
				{Key: dara.String(clusterTagKey), Value: dara.String(cfg.ClusterName)},
			},
		}, c.runtime)
	})
	if err != nil {
		return "", fmt.Errorf("CreateVSwitch: %w", err)
	}
	if response == nil || response.Body == nil || response.Body.VSwitchId == nil {
		return "", errors.New("CreateVSwitch returned no vSwitch ID")
	}
	return *response.Body.VSwitchId, nil
}

func validateVSwitch(item vSwitchInfo, zoneID string, requiredIPs int64) error {
	if item.ZoneID != zoneID {
		return fmt.Errorf("vSwitch %s is in zone %s, expected %s", item.ID, item.ZoneID, zoneID)
	}
	if item.Status != "Available" {
		return fmt.Errorf("vSwitch %s is not Available (status %s)", item.ID, item.Status)
	}
	if item.AvailableIP < requiredIPs {
		return fmt.Errorf("vSwitch %s has %d available IPs, need at least %d", item.ID, item.AvailableIP, requiredIPs)
	}
	if item.VPCID == "" {
		return fmt.Errorf("vSwitch %s is not attached to a VPC", item.ID)
	}
	return nil
}

func (c *Client) describeSecurityGroup(ctx context.Context, regionID, id string) (securityGroupInfo, error) {
	if err := ctx.Err(); err != nil {
		return securityGroupInfo{}, err
	}
	response, err := withCloudRetry(ctx, "DescribeSecurityGroups "+id, func() (*ecs.DescribeSecurityGroupsResponse, error) {
		return c.ecs.DescribeSecurityGroupsWithOptions(&ecs.DescribeSecurityGroupsRequest{
			RegionId: dara.String(regionID), SecurityGroupId: dara.String(id), MaxResults: dara.Int32(100),
		}, c.runtime)
	})
	if err != nil {
		return securityGroupInfo{}, fmt.Errorf("DescribeSecurityGroups %s: %w", id, err)
	}
	items := securityGroupItems(response)
	if len(items) != 1 || items[0].SecurityGroupId == nil {
		return securityGroupInfo{}, fmt.Errorf("security group %s was not found in region %s", id, regionID)
	}
	result := securityGroupInfo{
		ID: *items[0].SecurityGroupId, Name: stringValue(items[0].SecurityGroupName),
		VPCID: stringValue(items[0].VpcId), ServiceManaged: boolValue(items[0].ServiceManaged),
	}
	if result.VPCID == "" {
		return securityGroupInfo{}, fmt.Errorf("security group %s is not a VPC security group", id)
	}
	return result, nil
}

func (c *Client) findManagedSecurityGroup(ctx context.Context, cfg *config.Config, vpcID string) (securityGroupInfo, bool, error) {
	var candidates []securityGroupInfo
	var nextToken string
	for {
		if err := ctx.Err(); err != nil {
			return securityGroupInfo{}, false, err
		}
		request := &ecs.DescribeSecurityGroupsRequest{
			RegionId: dara.String(cfg.RegionID), VpcId: dara.String(vpcID),
			NetworkType: dara.String("vpc"), MaxResults: dara.Int32(100),
		}
		if nextToken != "" {
			request.NextToken = dara.String(nextToken)
		}
		response, err := withCloudRetry(ctx, "DescribeSecurityGroups managed", func() (*ecs.DescribeSecurityGroupsResponse, error) {
			return c.ecs.DescribeSecurityGroupsWithOptions(request, c.runtime)
		})
		if err != nil {
			return securityGroupInfo{}, false, fmt.Errorf(
				"find security group with prefix %s: %w", securityGroupNamePrefix, err,
			)
		}
		for _, item := range securityGroupItems(response) {
			if item != nil && item.SecurityGroupId != nil && stringValue(item.VpcId) == vpcID &&
				strings.HasPrefix(stringValue(item.SecurityGroupName), securityGroupNamePrefix) &&
				!boolValue(item.ServiceManaged) {
				candidates = append(candidates, securityGroupInfo{
					ID: *item.SecurityGroupId, Name: stringValue(item.SecurityGroupName),
					VPCID: vpcID, ServiceManaged: boolValue(item.ServiceManaged),
				})
			}
		}
		newToken := ""
		if response != nil && response.Body != nil {
			newToken = stringValue(response.Body.NextToken)
		}
		if newToken == "" || newToken == nextToken {
			break
		}
		nextToken = newToken
	}
	if len(candidates) == 0 {
		return securityGroupInfo{}, false, nil
	}
	preferredName := resourceName(securityGroupNamePrefix, cfg.ClusterName)
	sort.Slice(candidates, func(i, j int) bool {
		if (candidates[i].Name == preferredName) != (candidates[j].Name == preferredName) {
			return candidates[i].Name == preferredName
		}
		return candidates[i].ID < candidates[j].ID
	})
	return candidates[0], true, nil
}

func (c *Client) createSecurityGroup(ctx context.Context, cfg *config.Config, vpcID string, tokenTime time.Time) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	response, err := withCloudRetry(ctx, "CreateSecurityGroup", func() (*ecs.CreateSecurityGroupResponse, error) {
		return c.ecs.CreateSecurityGroupWithOptions(&ecs.CreateSecurityGroupRequest{
			RegionId:          dara.String(cfg.RegionID),
			VpcId:             dara.String(vpcID),
			SecurityGroupName: dara.String(resourceName(securityGroupNamePrefix, cfg.ClusterName)),
			SecurityGroupType: dara.String("normal"),
			Description:       dara.String("Ceph lab security group managed by CephTower"),
			ClientToken:       dara.String(clientToken(cfg.ClusterName, "security-group", tokenTime)),
			Tag: []*ecs.CreateSecurityGroupRequestTag{
				{Key: dara.String(managedByTagKey), Value: dara.String(managedByTagValue)},
				{Key: dara.String(clusterTagKey), Value: dara.String(cfg.ClusterName)},
			},
		}, c.runtime)
	})
	if err != nil {
		return "", fmt.Errorf("CreateSecurityGroup: %w", err)
	}
	if response == nil || response.Body == nil || response.Body.SecurityGroupId == nil {
		return "", errors.New("CreateSecurityGroup returned no security group ID")
	}
	return *response.Body.SecurityGroupId, nil
}

func (c *Client) authorizeLabSecurityGroup(ctx context.Context, cfg *config.Config, id string, tokenTime time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	_, err := withCloudRetry(ctx, "AuthorizeSecurityGroup "+id, func() (*ecs.AuthorizeSecurityGroupResponse, error) {
		return c.ecs.AuthorizeSecurityGroupWithOptions(&ecs.AuthorizeSecurityGroupRequest{
			RegionId: dara.String(cfg.RegionID), SecurityGroupId: dara.String(id),
			ClientToken: dara.String(clientToken(cfg.ClusterName, "security-rules", tokenTime)),
			Permissions: labSecurityGroupPermissions(cfg.Network.AccessSourceCIDR),
		}, c.runtime)
	})
	if err != nil {
		return fmt.Errorf("AuthorizeSecurityGroup %s: %w", id, err)
	}
	return nil
}

func labSecurityGroupPermissions(sourceCIDR string) []*ecs.AuthorizeSecurityGroupRequestPermissions {
	tcpRule := func(portRange, sourceCIDR, description string) *ecs.AuthorizeSecurityGroupRequestPermissions {
		return &ecs.AuthorizeSecurityGroupRequestPermissions{
			IpProtocol: dara.String("TCP"), NicType: dara.String("intranet"),
			PortRange: dara.String(portRange), SourceCidrIp: dara.String(sourceCIDR),
			Policy: dara.String("accept"), Priority: dara.String("1"),
			Description: dara.String(description),
		}
	}

	return []*ecs.AuthorizeSecurityGroupRequestPermissions{
		tcpRule("22/22", sourceCIDR, "SSH access from configured source CIDR"),
		tcpRule("8443/8443", sourceCIDR, "Ceph Dashboard HTTPS from configured source CIDR"),
		tcpRule("3300/3300", sourceCIDR, "Ceph monitor v2 from configured source CIDR"),
		tcpRule("6789/6789", sourceCIDR, "Ceph monitor v1 from configured source CIDR"),
		tcpRule("6800/7568", sourceCIDR, "Ceph daemon ports from configured source CIDR"),
		tcpRule("36900/36900", sourceCIDR, "CephTower HTTP service from configured source CIDR"),
	}
}

func (c *Client) waitVPCAvailable(ctx context.Context, regionID, id string) error {
	return waitFor(ctx, func() (bool, error) {
		response, err := withCloudRetry(ctx, "DescribeVpcs "+id, func() (*vpc.DescribeVpcsResponse, error) {
			return c.vpc.DescribeVpcsWithOptions(&vpc.DescribeVpcsRequest{
				RegionId: dara.String(regionID), VpcId: dara.String(id), PageSize: dara.Int32(50),
			}, c.runtime)
		})
		if err != nil {
			return false, err
		}
		items := vpcItems(response)
		return len(items) == 1 && stringValue(items[0].Status) == "Available", nil
	}, "VPC "+id+" to become Available")
}

func (c *Client) waitVSwitchAvailable(ctx context.Context, regionID, zoneID, id string) error {
	return waitFor(ctx, func() (bool, error) {
		response, err := withCloudRetry(ctx, "DescribeVSwitches "+id, func() (*vpc.DescribeVSwitchesResponse, error) {
			return c.vpc.DescribeVSwitchesWithOptions(&vpc.DescribeVSwitchesRequest{
				RegionId: dara.String(regionID), VSwitchId: dara.String(id), PageSize: dara.Int32(50),
			}, c.runtime)
		})
		if err != nil {
			return false, err
		}
		items := vSwitchItems(response)
		if len(items) != 1 || items[0].VSwitchId == nil {
			return false, nil
		}
		item := vSwitchInfo{
			ID: *items[0].VSwitchId, VPCID: stringValue(items[0].VpcId), ZoneID: stringValue(items[0].ZoneId),
			Status: stringValue(items[0].Status), AvailableIP: int64Value(items[0].AvailableIpAddressCount),
		}
		return item.Status == "Available" && item.ZoneID == zoneID && item.VPCID != "", nil
	}, "vSwitch "+id+" to become Available")
}

func (c *Client) waitVSwitchDeleted(ctx context.Context, regionID, id string) error {
	return waitFor(ctx, func() (bool, error) {
		response, err := withCloudRetry(ctx, "DescribeVSwitches "+id, func() (*vpc.DescribeVSwitchesResponse, error) {
			return c.vpc.DescribeVSwitchesWithOptions(&vpc.DescribeVSwitchesRequest{
				RegionId: dara.String(regionID), VSwitchId: dara.String(id), PageSize: dara.Int32(50),
			}, c.runtime)
		})
		if err != nil {
			return false, err
		}
		return len(vSwitchItems(response)) == 0, nil
	}, "vSwitch "+id+" to be deleted")
}

func waitFor(ctx context.Context, check func() (bool, error), description string) error {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	checks := 0
	logging.Infof("network: waiting for %s", description)
	for {
		checks++
		done, err := check()
		if err != nil {
			return err
		}
		if done {
			logging.Infof("network: %s completed after %d check(s)", description, checks)
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for %s: %w", description, ctx.Err())
		case <-ticker.C:
		}
	}
}

func vpcItems(response *vpc.DescribeVpcsResponse) []*vpc.DescribeVpcsResponseBodyVpcsVpc {
	if response == nil || response.Body == nil || response.Body.Vpcs == nil {
		return nil
	}
	return response.Body.Vpcs.Vpc
}

func vSwitchItems(response *vpc.DescribeVSwitchesResponse) []*vpc.DescribeVSwitchesResponseBodyVSwitchesVSwitch {
	if response == nil || response.Body == nil || response.Body.VSwitches == nil {
		return nil
	}
	return response.Body.VSwitches.VSwitch
}

func securityGroupItems(response *ecs.DescribeSecurityGroupsResponse) []*ecs.DescribeSecurityGroupsResponseBodySecurityGroupsSecurityGroup {
	if response == nil || response.Body == nil || response.Body.SecurityGroups == nil {
		return nil
	}
	return response.Body.SecurityGroups.SecurityGroup
}

func int64Value(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func int32Value(value *int32) int32 {
	if value == nil {
		return 0
	}
	return *value
}

func boolValue(value *bool) bool {
	return value != nil && *value
}

func resourceName(prefix, clusterName string) string {
	return prefix + clusterName
}

func cidrContains(parent, child string) bool {
	parentPrefix, parentErr := netip.ParsePrefix(parent)
	childPrefix, childErr := netip.ParsePrefix(child)
	return parentErr == nil && childErr == nil && parentPrefix.Addr().Is4() && childPrefix.Addr().Is4() &&
		childPrefix.Bits() >= parentPrefix.Bits() && parentPrefix.Contains(childPrefix.Addr())
}
