package command

import "context"

func (c *CommandClient) MonDump(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "mon", "dump")
}

func (c *CommandClient) MonMetadata(ctx context.Context, id string) (map[string]any, error) {
	return c.jsonObject(ctx, appendPositionals([]string{"mon", "metadata"}, id)...)
}

func (c *CommandClient) MonOKToStop(ctx context.Context, ids ...string) (map[string]any, error) {
	args := append([]string{"mon", "ok-to-stop"}, ids...)
	return c.jsonObject(ctx, args...)
}

func (c *CommandClient) MonVersions(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "mon", "versions")
}

func (c *CommandClient) MonScrub(ctx context.Context) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: []string{"mon", "scrub"}})
}

func (c *CommandClient) MonRemove(ctx context.Context, name string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: appendPositionals([]string{"mon", "rm"}, name)})
}

func (c *CommandClient) MgrDump(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "mgr", "dump")
}

func (c *CommandClient) MgrStat(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "mgr", "stat")
}

func (c *CommandClient) MgrMetadata(ctx context.Context, id string) (map[string]any, error) {
	return c.jsonObject(ctx, appendPositionals([]string{"mgr", "metadata"}, id)...)
}

func (c *CommandClient) MgrModuleList(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "mgr", "module", "ls")
}

func (c *CommandClient) MgrModuleEnable(ctx context.Context, module string, force bool) (CommandResult, error) {
	args := appendPositionals([]string{"mgr", "module", "enable"}, module)
	args = appendBoolFlag(args, "--force", force)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) MgrModuleDisable(ctx context.Context, module string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: appendPositionals([]string{"mgr", "module", "disable"}, module)})
}

func (c *CommandClient) MgrFail(ctx context.Context, mgrID string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: appendPositionals([]string{"mgr", "fail"}, mgrID)})
}

func (c *CommandClient) MgrSetDown(ctx context.Context, yes bool) (CommandResult, error) {
	args := []string{"mgr", "set", "down"}
	args = appendBoolFlag(args, "--yes-i-really-mean-it", yes)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) MgrVersions(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "mgr", "versions")
}
