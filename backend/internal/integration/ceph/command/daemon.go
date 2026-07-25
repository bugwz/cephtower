package command

import "context"

type OrchPSOptions struct {
	Hostname    string
	DaemonType  string
	DaemonID    string
	ServiceName string
	SortBy      string
	Refresh     bool
}

func (c *CommandClient) OrchPS(ctx context.Context, options OrchPSOptions) ([]map[string]any, error) {
	args := []string{"orch", "ps"}
	if options.Hostname != "" {
		args = append(args, options.Hostname)
	}
	if options.DaemonType != "" {
		args = append(args, "--daemon-type", options.DaemonType)
	}
	if options.DaemonID != "" {
		args = append(args, "--daemon-id", options.DaemonID)
	}
	if options.ServiceName != "" {
		args = append(args, "--service-name", options.ServiceName)
	}
	if options.SortBy != "" {
		args = append(args, "--sort-by", options.SortBy)
	}
	if options.Refresh {
		args = append(args, "--refresh")
	}

	var payload []map[string]any
	err := c.JSON(ctx, CommandRequest{Args: args}, &payload)
	return payload, err
}

func (c *CommandClient) MgrServices(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "mgr", "services")
}

func (c *CommandClient) MonStat(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "mon", "stat")
}

func (c *CommandClient) Tell(ctx context.Context, daemon string, command ...string) (map[string]any, error) {
	args := append([]string{"tell", daemon}, command...)
	return c.jsonObject(ctx, args...)
}

func (c *CommandClient) Daemon(ctx context.Context, daemon string, command ...string) (map[string]any, error) {
	args := append([]string{"daemon", daemon}, command...)
	return c.jsonObject(ctx, args...)
}
