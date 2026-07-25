package command

import "context"

type OrchHostListOptions struct {
	HostPattern string
	Label       string
	HostStatus  string
	Detail      bool
}

func (c *CommandClient) OrchHosts(ctx context.Context, options OrchHostListOptions) ([]map[string]any, error) {
	args := []string{"orch", "host", "ls"}
	if options.HostPattern != "" {
		args = append(args, "--host-pattern", options.HostPattern)
	}
	if options.Label != "" {
		args = append(args, "--label", options.Label)
	}
	if options.HostStatus != "" {
		args = append(args, "--host-status", options.HostStatus)
	}
	if options.Detail {
		args = append(args, "--detail")
	}

	var payload []map[string]any
	err := c.JSON(ctx, CommandRequest{Args: args}, &payload)
	return payload, err
}

func (c *CommandClient) OrchHostAdd(ctx context.Context, hostname string, addr string, maintenance bool, labels ...string) (CommandResult, error) {
	args := appendPositionals([]string{"orch", "host", "add"}, hostname, addr)
	args = append(args, labels...)
	args = appendBoolFlag(args, "--maintenance", maintenance)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) OrchHostRemove(ctx context.Context, hostname string, force bool, offline bool, removeCrushEntry bool) (CommandResult, error) {
	args := appendPositionals([]string{"orch", "host", "rm"}, hostname)
	args = appendBoolFlag(args, "--force", force)
	args = appendBoolFlag(args, "--offline", offline)
	args = appendBoolFlag(args, "--rm-crush-entry", removeCrushEntry)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) OrchHostSetAddr(ctx context.Context, hostname string, addr string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: appendPositionals([]string{"orch", "host", "set-addr"}, hostname, addr)})
}

func (c *CommandClient) OrchHostLabelAdd(ctx context.Context, hostname string, label string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: appendPositionals([]string{"orch", "host", "label", "add"}, hostname, label)})
}

func (c *CommandClient) OrchHostLabelRemove(ctx context.Context, hostname string, label string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: appendPositionals([]string{"orch", "host", "label", "rm"}, hostname, label)})
}

func (c *CommandClient) OrchHostDrain(ctx context.Context, hostname string, force bool, keepConfKeyring bool, zapOSDDevices bool) (CommandResult, error) {
	args := appendPositionals([]string{"orch", "host", "drain"}, hostname)
	args = appendBoolFlag(args, "--force", force)
	args = appendBoolFlag(args, "--keep-conf-keyring", keepConfKeyring)
	args = appendBoolFlag(args, "--zap-osd-devices", zapOSDDevices)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) OrchHostDrainStop(ctx context.Context, hostname string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: appendPositionals([]string{"orch", "host", "drain", "stop"}, hostname)})
}

func (c *CommandClient) OrchHostMaintenanceEnter(ctx context.Context, hostname string, force bool, yes bool) (CommandResult, error) {
	args := appendPositionals([]string{"orch", "host", "maintenance", "enter"}, hostname)
	args = appendBoolFlag(args, "--force", force)
	args = appendBoolFlag(args, "--yes-i-really-mean-it", yes)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) OrchHostMaintenanceExit(ctx context.Context, hostname string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: appendPositionals([]string{"orch", "host", "maintenance", "exit"}, hostname)})
}

func (c *CommandClient) OrchHostOKToStop(ctx context.Context, hostname string) (map[string]any, error) {
	return c.jsonObject(ctx, appendPositionals([]string{"orch", "host", "ok-to-stop"}, hostname)...)
}

func (c *CommandClient) OrchHostRescan(ctx context.Context, hostname string, withSummary bool) (CommandResult, error) {
	args := appendPositionals([]string{"orch", "host", "rescan"}, hostname)
	args = appendBoolFlag(args, "--with-summary", withSummary)
	return c.Run(ctx, CommandRequest{Args: args})
}
