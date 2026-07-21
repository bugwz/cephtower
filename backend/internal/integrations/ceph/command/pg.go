package command

import "context"

func (c *CommandClient) PGDump(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "pg", "dump")
}

func (c *CommandClient) PGDumpPgsBrief(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "pg", "dump", "pgs_brief")
}

func (c *CommandClient) PGDumpJSON(ctx context.Context, contents ...string) (map[string]any, error) {
	return c.jsonObject(ctx, append([]string{"pg", "dump_json"}, contents...)...)
}

func (c *CommandClient) PGDumpPoolsJSON(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "pg", "dump_pools_json")
}

func (c *CommandClient) PGDumpStuck(ctx context.Context, stuckType string, threshold string) (map[string]any, error) {
	args := []string{"pg", "dump_stuck"}
	args = appendPositionals(args, stuckType, threshold)
	return c.jsonObject(ctx, args...)
}

func (c *CommandClient) PGMap(ctx context.Context, pgid string) (map[string]any, error) {
	return c.jsonObject(ctx, appendPositionals([]string{"pg", "map"}, pgid)...)
}

func (c *CommandClient) PGScrub(ctx context.Context, pgid string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: appendPositionals([]string{"pg", "scrub"}, pgid)})
}

func (c *CommandClient) PGDeepScrub(ctx context.Context, pgid string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: appendPositionals([]string{"pg", "deep-scrub"}, pgid)})
}

func (c *CommandClient) PGRepair(ctx context.Context, pgid string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: appendPositionals([]string{"pg", "repair"}, pgid)})
}

func (c *CommandClient) PGRepeer(ctx context.Context, pgid string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: appendPositionals([]string{"pg", "repeer"}, pgid)})
}

func (c *CommandClient) PGForceRecovery(ctx context.Context, pgids ...string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: append([]string{"pg", "force-recovery"}, pgids...)})
}

func (c *CommandClient) PGForceBackfill(ctx context.Context, pgids ...string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: append([]string{"pg", "force-backfill"}, pgids...)})
}

func (c *CommandClient) PGCancelForceRecovery(ctx context.Context, pgids ...string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: append([]string{"pg", "cancel-force-recovery"}, pgids...)})
}

func (c *CommandClient) PGCancelForceBackfill(ctx context.Context, pgids ...string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: append([]string{"pg", "cancel-force-backfill"}, pgids...)})
}
