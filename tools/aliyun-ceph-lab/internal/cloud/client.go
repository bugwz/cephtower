package cloud

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	openapiutil "github.com/alibabacloud-go/darabonba-openapi/v2/utils"
	ecs "github.com/alibabacloud-go/ecs-20140526/v7/client"
	"github.com/alibabacloud-go/tea/dara"
	vpc "github.com/alibabacloud-go/vpc-20160428/v6/client"

	"cephtower/tools/aliyun-ceph-lab/internal/config"
)

type Client struct {
	ecs     *ecs.Client
	vpc     *vpc.Client
	runtime *dara.RuntimeOptions
}

type Instance struct {
	ID              string
	Name            string
	Status          string
	PublicIP        string
	PrivateIP       string
	AutoReleaseTime string
}

func New(cfg *config.Config) (*Client, error) {
	if err := cfg.ValidateCloudCredentials(); err != nil {
		return nil, err
	}

	newSDKConfig := func(endpoint string) *openapiutil.Config {
		value := &openapiutil.Config{
			AccessKeyId:     dara.String(cfg.AccessKeyID),
			AccessKeySecret: dara.String(cfg.AccessKeySecret),
			RegionId:        dara.String(cfg.RegionID),
		}
		if cfg.SecurityToken != "" {
			value.SecurityToken = dara.String(cfg.SecurityToken)
		}
		if endpoint != "" {
			value.Endpoint = dara.String(endpoint)
		}
		return value
	}
	ecsClient, err := ecs.NewClient(newSDKConfig(cfg.Endpoint))
	if err != nil {
		return nil, fmt.Errorf("create ECS client: %w", err)
	}
	vpcClient, err := vpc.NewClient(newSDKConfig(cfg.VPCEndpoint))
	if err != nil {
		return nil, fmt.Errorf("create VPC client: %w", err)
	}
	runtime := (&dara.RuntimeOptions{}).
		SetAutoretry(true).
		SetMaxAttempts(3).
		SetBackoffPolicy("no").
		SetConnectTimeout(10_000).
		SetReadTimeout(30_000)
	return &Client{ecs: ecsClient, vpc: vpcClient, runtime: runtime}, nil
}

func (c *Client) RunNode(cfg *config.Config, node config.Node, expiresAt time.Time) (string, error) {
	dataDisks := make([]*ecs.RunInstancesRequestDataDisk, 0, len(node.DataDisks))
	for _, disk := range node.DataDisks {
		dataDisk := &ecs.RunInstancesRequestDataDisk{
			Category:           dara.String(disk.Category),
			Size:               dara.Int32(disk.SizeGiB),
			DeleteWithInstance: dara.Bool(true),
		}
		if disk.PerformanceLevel != "" {
			dataDisk.PerformanceLevel = dara.String(disk.PerformanceLevel)
		}
		dataDisks = append(dataDisks, dataDisk)
	}
	systemDisk := &ecs.RunInstancesRequestSystemDisk{
		Category: dara.String(cfg.SystemDiskCategory),
		Size:     dara.String(strconv.FormatInt(int64(cfg.SystemDiskSizeGiB), 10)),
	}
	if cfg.SystemDiskPerformanceLevel != "" {
		systemDisk.PerformanceLevel = dara.String(cfg.SystemDiskPerformanceLevel)
	}
	request := &ecs.RunInstancesRequest{
		RegionId:                dara.String(cfg.RegionID),
		ZoneId:                  dara.String(cfg.ZoneID),
		VSwitchId:               dara.String(cfg.VSwitchID),
		SecurityGroupId:         dara.String(cfg.SecurityGroupID),
		ImageId:                 dara.String(cfg.ImageID),
		Password:                dara.String(cfg.SSHPassword),
		InstanceType:            dara.String(node.InstanceType),
		InstanceName:            dara.String(node.Name),
		HostName:                dara.String(node.Name),
		InstanceChargeType:      dara.String("PostPaid"),
		Amount:                  dara.Int32(1),
		MinAmount:               dara.Int32(1),
		ClientToken:             dara.String(clientToken(cfg.ClusterName, node.Name, expiresAt)),
		AutoReleaseTime:         dara.String(expiresAt.UTC().Truncate(time.Minute).Format("2006-01-02T15:04:00Z")),
		DeletionProtection:      dara.Bool(false),
		InternetChargeType:      dara.String(cfg.InternetChargeType),
		InternetMaxBandwidthOut: dara.Int32(cfg.InternetMaxBandwidthOutMbps),
		SystemDisk:              systemDisk,
		DataDisk:                dataDisks,
		Tag: []*ecs.RunInstancesRequestTag{
			{Key: dara.String("ceph-lab-cluster"), Value: dara.String(cfg.ClusterName)},
			{Key: dara.String("ceph-lab-managed-by"), Value: dara.String("cephtower")},
		},
	}
	if cfg.CPUOptions != nil {
		request.CpuOptions = &ecs.RunInstancesRequestCpuOptions{
			Core:           dara.Int32(cfg.CPUOptions.CoreCount),
			ThreadsPerCore: dara.Int32(cfg.CPUOptions.ThreadsPerCore),
		}
	}
	if cfg.HTTPTokens != "" {
		request.HttpTokens = dara.String(cfg.HTTPTokens)
	}
	if cfg.SecurityEnhancementStrategy != "" {
		request.SecurityEnhancementStrategy = dara.String(cfg.SecurityEnhancementStrategy)
	}
	response, err := c.ecs.RunInstancesWithOptions(request, c.runtime)
	if err != nil {
		return "", fmt.Errorf("RunInstances for %s: %w", node.Name, err)
	}
	if response == nil || response.Body == nil || response.Body.InstanceIdSets == nil ||
		len(response.Body.InstanceIdSets.InstanceIdSet) != 1 || response.Body.InstanceIdSets.InstanceIdSet[0] == nil {
		return "", fmt.Errorf("RunInstances for %s returned no instance ID", node.Name)
	}
	return *response.Body.InstanceIdSets.InstanceIdSet[0], nil
}

func clientToken(clusterName, nodeName string, expiresAt time.Time) string {
	sum := sha256.Sum256([]byte(clusterName + "\x00" + nodeName + "\x00" + expiresAt.UTC().Format(time.RFC3339)))
	return fmt.Sprintf("ceph-lab-%x", sum[:16])
}

func (c *Client) Describe(regionID string, instanceIDs []string) ([]Instance, error) {
	if len(instanceIDs) == 0 {
		return nil, nil
	}
	encoded, err := json.Marshal(instanceIDs)
	if err != nil {
		return nil, fmt.Errorf("encode instance IDs: %w", err)
	}
	response, err := c.ecs.DescribeInstancesWithOptions(&ecs.DescribeInstancesRequest{
		RegionId:    dara.String(regionID),
		InstanceIds: dara.String(string(encoded)),
		PageSize:    dara.Int32(100),
	}, c.runtime)
	if err != nil {
		return nil, fmt.Errorf("DescribeInstances: %w", err)
	}
	if response == nil || response.Body == nil || response.Body.Instances == nil {
		return nil, nil
	}
	result := make([]Instance, 0, len(response.Body.Instances.Instance))
	for _, item := range response.Body.Instances.Instance {
		if item == nil || item.InstanceId == nil {
			continue
		}
		instance := Instance{ID: *item.InstanceId}
		instance.Name = stringValue(item.InstanceName)
		instance.Status = stringValue(item.Status)
		instance.AutoReleaseTime = stringValue(item.AutoReleaseTime)
		if item.PublicIpAddress != nil {
			instance.PublicIP = firstString(item.PublicIpAddress.IpAddress)
		}
		if item.VpcAttributes != nil && item.VpcAttributes.PrivateIpAddress != nil {
			instance.PrivateIP = firstString(item.VpcAttributes.PrivateIpAddress.IpAddress)
		}
		result = append(result, instance)
	}
	return result, nil
}

func (c *Client) Delete(instanceID string) error {
	_, err := c.ecs.DeleteInstanceWithOptions(&ecs.DeleteInstanceRequest{
		InstanceId: dara.String(instanceID),
		Force:      dara.Bool(true),
		ForceStop:  dara.Bool(false),
	}, c.runtime)
	if err != nil {
		return fmt.Errorf("DeleteInstance %s: %w", instanceID, err)
	}
	return nil
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func firstString(values []*string) string {
	for _, value := range values {
		if value != nil && *value != "" {
			return *value
		}
	}
	return ""
}
