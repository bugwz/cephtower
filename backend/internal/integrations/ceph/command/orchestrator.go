package command

import "context"

type OrchListOptions struct {
	ServiceType string
	ServiceName string
	Export      bool
	Refresh     bool
}

type OrchDeviceListOptions struct {
	Hostname string
	Refresh  bool
	Wide     bool
	Summary  bool
}

type OrchApplyOptions struct {
	ServiceType     string
	Placement       string
	Unmanaged       bool
	DryRun          bool
	NoOverwrite     bool
	ContinueOnError bool
	SpecInput       []byte
}

func (c *CommandClient) OrchStatus(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "orch", "status")
}

func (c *CommandClient) OrchList(ctx context.Context, options OrchListOptions) ([]map[string]any, error) {
	args := []string{"orch", "ls"}
	args = appendStringFlag(args, "--service-type", options.ServiceType)
	args = appendStringFlag(args, "--service-name", options.ServiceName)
	args = appendBoolFlag(args, "--export", options.Export)
	args = appendBoolFlag(args, "--refresh", options.Refresh)
	var payload []map[string]any
	err := c.JSON(ctx, CommandRequest{Args: args}, &payload)
	return payload, err
}

func (c *CommandClient) OrchPause(ctx context.Context) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: []string{"orch", "pause"}})
}

func (c *CommandClient) OrchResume(ctx context.Context) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: []string{"orch", "resume"}})
}

func (c *CommandClient) OrchCancel(ctx context.Context) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: []string{"orch", "cancel"}})
}

func (c *CommandClient) OrchRemoveService(ctx context.Context, serviceName string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: appendPositionals([]string{"orch", "rm"}, serviceName)})
}

func (c *CommandClient) OrchDaemonAction(ctx context.Context, action string, daemonName string, force bool) (CommandResult, error) {
	args := appendPositionals([]string{"orch", "daemon"}, action, daemonName)
	args = appendBoolFlag(args, "--force", force)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) OrchDaemonRemove(ctx context.Context, names []string, force bool) (CommandResult, error) {
	args := append([]string{"orch", "daemon", "rm"}, names...)
	args = appendBoolFlag(args, "--force", force)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) OrchDaemonRedeploy(ctx context.Context, names []string, image string, force bool) (CommandResult, error) {
	args := append([]string{"orch", "daemon", "redeploy"}, names...)
	args = appendStringFlag(args, "--image", image)
	args = appendBoolFlag(args, "--force", force)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) OrchApply(ctx context.Context, options OrchApplyOptions) (CommandResult, error) {
	args := []string{"orch", "apply"}
	if len(options.SpecInput) > 0 {
		args = append(args, "-i", "-")
	} else {
		args = appendPositionals(args, options.ServiceType, options.Placement)
	}
	args = appendBoolFlag(args, "--unmanaged", options.Unmanaged)
	args = appendBoolFlag(args, "--dry-run", options.DryRun)
	args = appendBoolFlag(args, "--no-overwrite", options.NoOverwrite)
	args = appendBoolFlag(args, "--continue-on-error", options.ContinueOnError)
	return c.Run(ctx, CommandRequest{Args: args, Input: options.SpecInput})
}

func (c *CommandClient) OrchApplyMDS(ctx context.Context, fsName string, placement string, dryRun bool, unmanaged bool) (CommandResult, error) {
	args := appendPositionals([]string{"orch", "apply", "mds"}, fsName, placement)
	args = appendBoolFlag(args, "--dry-run", dryRun)
	args = appendBoolFlag(args, "--unmanaged", unmanaged)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) OrchApplyRGW(ctx context.Context, serviceID string, realmName string, zoneName string, placement string, dryRun bool, unmanaged bool) (CommandResult, error) {
	args := appendPositionals([]string{"orch", "apply", "rgw"}, serviceID, placement)
	args = appendStringFlag(args, "--realm", realmName)
	args = appendStringFlag(args, "--zone", zoneName)
	args = appendBoolFlag(args, "--dry-run", dryRun)
	args = appendBoolFlag(args, "--unmanaged", unmanaged)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) OrchApplyOSD(ctx context.Context, allAvailableDevices bool, method string, dryRun bool, unmanaged bool) (CommandResult, error) {
	args := []string{"orch", "apply", "osd"}
	args = appendBoolFlag(args, "--all-available-devices", allAvailableDevices)
	args = appendStringFlag(args, "--method", method)
	args = appendBoolFlag(args, "--dry-run", dryRun)
	args = appendBoolFlag(args, "--unmanaged", unmanaged)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) OrchDeviceList(ctx context.Context, options OrchDeviceListOptions) ([]map[string]any, error) {
	args := appendPositionals([]string{"orch", "device", "ls"}, options.Hostname)
	args = appendBoolFlag(args, "--refresh", options.Refresh)
	args = appendBoolFlag(args, "--wide", options.Wide)
	args = appendBoolFlag(args, "--summary", options.Summary)
	var payload []map[string]any
	err := c.JSON(ctx, CommandRequest{Args: args}, &payload)
	return payload, err
}

func (c *CommandClient) OrchDeviceZap(ctx context.Context, hostname string, path string, force bool) (CommandResult, error) {
	args := appendPositionals([]string{"orch", "device", "zap"}, hostname, path)
	args = appendBoolFlag(args, "--force", force)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) OrchOSDRemove(ctx context.Context, osdIDs []string, replace bool, force bool, zap bool) (CommandResult, error) {
	args := append([]string{"orch", "osd", "rm"}, osdIDs...)
	args = appendBoolFlag(args, "--replace", replace)
	args = appendBoolFlag(args, "--force", force)
	args = appendBoolFlag(args, "--zap", zap)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) OrchOSDRemoveStatus(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "orch", "osd", "rm", "status")
}

func (c *CommandClient) OrchUpgradeCheck(ctx context.Context, image string, version string) (map[string]any, error) {
	args := []string{"orch", "upgrade", "check"}
	args = appendStringFlag(args, "--image", image)
	args = appendStringFlag(args, "--ceph-version", version)
	return c.jsonObject(ctx, args...)
}

func (c *CommandClient) OrchUpgradeList(ctx context.Context, image string, tags bool, showAllVersions bool) (map[string]any, error) {
	args := []string{"orch", "upgrade", "ls"}
	args = appendStringFlag(args, "--image", image)
	args = appendBoolFlag(args, "--tags", tags)
	args = appendBoolFlag(args, "--show-all-versions", showAllVersions)
	return c.jsonObject(ctx, args...)
}

func (c *CommandClient) OrchUpgradeStatus(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "orch", "upgrade", "status")
}

func (c *CommandClient) OrchUpgradeStart(ctx context.Context, image string, version string, serviceTypes []string, hosts []string, services []string, limit string) (CommandResult, error) {
	args := appendPositionals([]string{"orch", "upgrade", "start"}, image)
	args = appendStringFlag(args, "--ceph-version", version)
	args = appendRepeatedFlag(args, "--daemon-types", serviceTypes)
	args = appendRepeatedFlag(args, "--hosts", hosts)
	args = appendRepeatedFlag(args, "--services", services)
	args = appendStringFlag(args, "--limit", limit)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) OrchUpgradePause(ctx context.Context) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: []string{"orch", "upgrade", "pause"}})
}

func (c *CommandClient) OrchUpgradeResume(ctx context.Context) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: []string{"orch", "upgrade", "resume"}})
}

func (c *CommandClient) OrchUpgradeStop(ctx context.Context) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: []string{"orch", "upgrade", "stop"}})
}
