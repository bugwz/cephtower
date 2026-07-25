package command

import "context"

func (c *CommandClient) ServiceDump(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "service", "dump")
}

func (c *CommandClient) ServiceStatus(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "service", "status")
}

func (c *CommandClient) DeviceList(ctx context.Context) ([]map[string]any, error) {
	var payload []map[string]any
	err := c.JSON(ctx, CommandRequest{Args: []string{"device", "ls"}}, &payload)
	return payload, err
}

func (c *CommandClient) DeviceInfo(ctx context.Context, devid string) (map[string]any, error) {
	return c.jsonObject(ctx, appendPositionals([]string{"device", "info"}, devid)...)
}

func (c *CommandClient) DeviceListByHost(ctx context.Context, hostname string) ([]map[string]any, error) {
	var payload []map[string]any
	err := c.JSON(ctx, CommandRequest{Args: appendPositionals([]string{"device", "ls-by-host"}, hostname)}, &payload)
	return payload, err
}

func (c *CommandClient) DeviceListByDaemon(ctx context.Context, daemon string) ([]map[string]any, error) {
	var payload []map[string]any
	err := c.JSON(ctx, CommandRequest{Args: appendPositionals([]string{"device", "ls-by-daemon"}, daemon)}, &payload)
	return payload, err
}

func (c *CommandClient) DeviceHealthMetrics(ctx context.Context, devid string) (map[string]any, error) {
	return c.jsonObject(ctx, appendPositionals([]string{"device", "get-health-metrics"}, devid)...)
}

func (c *CommandClient) DeviceCheckHealth(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "device", "check-health")
}

func (c *CommandClient) DeviceMonitoringOn(ctx context.Context) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: []string{"device", "monitoring", "on"}})
}

func (c *CommandClient) DeviceMonitoringOff(ctx context.Context) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: []string{"device", "monitoring", "off"}})
}

func (c *CommandClient) BalancerStatus(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "balancer", "status")
}

func (c *CommandClient) BalancerDump(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "balancer", "dump")
}

func (c *CommandClient) BalancerMode(ctx context.Context, mode string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: appendPositionals([]string{"balancer", "mode"}, mode)})
}

func (c *CommandClient) BalancerOn(ctx context.Context) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: []string{"balancer", "on"}})
}

func (c *CommandClient) BalancerOff(ctx context.Context) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: []string{"balancer", "off"}})
}

func (c *CommandClient) Progress(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "progress")
}

func (c *CommandClient) ProgressJSON(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "progress", "json")
}

func (c *CommandClient) ProgressClear(ctx context.Context) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: []string{"progress", "clear"}})
}

func (c *CommandClient) NFSClusterList(ctx context.Context) ([]map[string]any, error) {
	var payload []map[string]any
	err := c.JSON(ctx, CommandRequest{Args: []string{"nfs", "cluster", "ls"}}, &payload)
	return payload, err
}

func (c *CommandClient) NFSClusterInfo(ctx context.Context, clusterID string) (map[string]any, error) {
	return c.jsonObject(ctx, appendPositionals([]string{"nfs", "cluster", "info"}, clusterID)...)
}

func (c *CommandClient) NFSExportList(ctx context.Context, clusterID string, detailed bool) ([]map[string]any, error) {
	args := appendPositionals([]string{"nfs", "export", "ls"}, clusterID)
	args = appendBoolFlag(args, "--detailed", detailed)
	var payload []map[string]any
	err := c.JSON(ctx, CommandRequest{Args: args}, &payload)
	return payload, err
}

func (c *CommandClient) NFSExportGet(ctx context.Context, clusterID string, pseudoPath string) (map[string]any, error) {
	return c.jsonObject(ctx, appendPositionals([]string{"nfs", "export", "get"}, clusterID, pseudoPath)...)
}

func (c *CommandClient) SMBClusterList(ctx context.Context) ([]map[string]any, error) {
	var payload []map[string]any
	err := c.JSON(ctx, CommandRequest{Args: []string{"smb", "cluster", "ls"}}, &payload)
	return payload, err
}

func (c *CommandClient) SMBShareList(ctx context.Context, clusterID string) ([]map[string]any, error) {
	var payload []map[string]any
	err := c.JSON(ctx, CommandRequest{Args: appendPositionals([]string{"smb", "share", "ls"}, clusterID)}, &payload)
	return payload, err
}

func (c *CommandClient) SMBShow(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "smb", "show")
}

func (c *CommandClient) RBDTaskList(ctx context.Context, taskID string) (map[string]any, error) {
	return c.jsonObject(ctx, appendPositionals([]string{"rbd", "task", "list"}, taskID)...)
}

func (c *CommandClient) RBDTaskCancel(ctx context.Context, taskID string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: appendPositionals([]string{"rbd", "task", "cancel"}, taskID)})
}

func (c *CommandClient) RBDPerfImageCounters(ctx context.Context, poolSpec string, sortBy string) (map[string]any, error) {
	args := appendPositionals([]string{"rbd", "perf", "image", "counters"}, poolSpec)
	args = appendStringFlag(args, "--sort-by", sortBy)
	return c.jsonObject(ctx, args...)
}

func (c *CommandClient) RBDPerfImageStats(ctx context.Context, poolSpec string, sortBy string) (map[string]any, error) {
	args := appendPositionals([]string{"rbd", "perf", "image", "stats"}, poolSpec)
	args = appendStringFlag(args, "--sort-by", sortBy)
	return c.jsonObject(ctx, args...)
}
