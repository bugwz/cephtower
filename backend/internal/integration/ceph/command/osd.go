package command

import (
	"context"
	"strconv"
)

type OSDPoolCreateOptions struct {
	PGNum              *int
	PGPNum             *int
	PoolType           string
	ErasureCodeProfile string
	Rule               string
	ExpectedNumObjects *int
	Size               *int
	PGNumMin           *int
	PGNumMax           *int
	AutoscaleMode      string
	Bulk               bool
	TargetSizeBytes    string
	TargetSizeRatio    *float64
	Yes                bool
}

func (c *CommandClient) OSDTree(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "osd", "tree")
}

func (c *CommandClient) OSDDF(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "osd", "df")
}

func (c *CommandClient) OSDDump(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "osd", "dump")
}

func (c *CommandClient) OSDStat(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "osd", "stat")
}

func (c *CommandClient) OSDMap(ctx context.Context, pool string, object string, namespace string) (map[string]any, error) {
	args := appendPositionals([]string{"osd", "map"}, pool, object)
	args = appendStringFlag(args, "--namespace", namespace)
	return c.jsonObject(ctx, args...)
}

func (c *CommandClient) OSDMetadata(ctx context.Context, id string) (map[string]any, error) {
	return c.jsonObject(ctx, appendPositionals([]string{"osd", "metadata"}, id)...)
}

func (c *CommandClient) OSDPerf(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "osd", "perf")
}

func (c *CommandClient) OSDPoolList(ctx context.Context, detail bool) ([]map[string]any, error) {
	args := []string{"osd", "pool", "ls"}
	args = appendBoolFlag(args, "detail", detail)
	var payload []map[string]any
	err := c.JSON(ctx, CommandRequest{Args: args}, &payload)
	return payload, err
}

func (c *CommandClient) OSDPoolStats(ctx context.Context, pool string) (map[string]any, error) {
	return c.jsonObject(ctx, appendPositionals([]string{"osd", "pool", "stats"}, pool)...)
}

func (c *CommandClient) OSDPoolGet(ctx context.Context, pool string, variable string) (map[string]any, error) {
	return c.jsonObject(ctx, appendPositionals([]string{"osd", "pool", "get"}, pool, variable)...)
}

func (c *CommandClient) OSDPoolSet(ctx context.Context, pool string, variable string, value string, yes bool) (CommandResult, error) {
	args := appendPositionals([]string{"osd", "pool", "set"}, pool, variable, value)
	args = appendBoolFlag(args, "--yes-i-really-mean-it", yes)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) OSDPoolCreate(ctx context.Context, pool string, pgNum *int, pgpNum *int) (CommandResult, error) {
	return c.OSDPoolCreateWithOptions(ctx, pool, OSDPoolCreateOptions{PGNum: pgNum, PGPNum: pgpNum})
}

func (c *CommandClient) OSDPoolCreateWithOptions(ctx context.Context, pool string, options OSDPoolCreateOptions) (CommandResult, error) {
	args := appendPositionals([]string{"osd", "pool", "create"}, pool)
	positionals := []string{
		intValue(options.PGNum),
		intValue(options.PGPNum),
		options.PoolType,
		options.ErasureCodeProfile,
		options.Rule,
		intValue(options.ExpectedNumObjects),
		intValue(options.Size),
		intValue(options.PGNumMin),
		intValue(options.PGNumMax),
		options.AutoscaleMode,
	}
	args = appendSequentialPositionals(args, positionals...)
	args = appendBoolFlag(args, "--bulk", options.Bulk)
	args = appendStringFlag(args, "--target-size-bytes", options.TargetSizeBytes)
	args = appendFloatFlag(args, "--target-size-ratio", options.TargetSizeRatio)
	args = appendBoolFlag(args, "--yes-i-really-mean-it", options.Yes)
	return c.Run(ctx, CommandRequest{Args: args})
}

func intValue(value *int) string {
	if value == nil {
		return ""
	}
	return strconv.Itoa(*value)
}

func (c *CommandClient) OSDPoolRemove(ctx context.Context, pool string, yes bool) (CommandResult, error) {
	args := appendPositionals([]string{"osd", "pool", "rm"}, pool, pool)
	args = appendBoolFlag(args, "--yes-i-really-really-mean-it", yes)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) OSDPoolRename(ctx context.Context, srcPool string, destPool string, yes bool) (CommandResult, error) {
	args := appendPositionals([]string{"osd", "pool", "rename"}, srcPool, destPool)
	args = appendBoolFlag(args, "--yes-i-really-mean-it", yes)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) OSDPoolMakeSnapshot(ctx context.Context, pool string, snapshot string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: appendPositionals([]string{"osd", "pool", "mksnap"}, pool, snapshot)})
}

func (c *CommandClient) OSDPoolRemoveSnapshot(ctx context.Context, pool string, snapshot string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: appendPositionals([]string{"osd", "pool", "rmsnap"}, pool, snapshot)})
}

func (c *CommandClient) OSDCrushTree(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "osd", "crush", "tree")
}

func (c *CommandClient) OSDCrushDump(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "osd", "crush", "dump")
}

func (c *CommandClient) OSDCrushRuleList(ctx context.Context) ([]map[string]any, error) {
	var payload []map[string]any
	err := c.JSON(ctx, CommandRequest{Args: []string{"osd", "crush", "rule", "ls"}}, &payload)
	return payload, err
}

func (c *CommandClient) OSDCrushRuleDump(ctx context.Context, rule string) (map[string]any, error) {
	return c.jsonObject(ctx, appendPositionals([]string{"osd", "crush", "rule", "dump"}, rule)...)
}

func (c *CommandClient) OSDCrushClassList(ctx context.Context) ([]string, error) {
	var payload []string
	err := c.JSON(ctx, CommandRequest{Args: []string{"osd", "crush", "class", "ls"}}, &payload)
	return payload, err
}

func (c *CommandClient) OSDCrushClassListOSD(ctx context.Context, class string) ([]int, error) {
	var payload []int
	err := c.JSON(ctx, CommandRequest{Args: appendPositionals([]string{"osd", "crush", "class", "ls-osd"}, class)}, &payload)
	return payload, err
}

func (c *CommandClient) OSDSet(ctx context.Context, flag string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: appendPositionals([]string{"osd", "set"}, flag)})
}

func (c *CommandClient) OSDUnset(ctx context.Context, flag string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: appendPositionals([]string{"osd", "unset"}, flag)})
}

func (c *CommandClient) OSDMarkIn(ctx context.Context, ids ...string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: append([]string{"osd", "in"}, ids...)})
}

func (c *CommandClient) OSDMarkOut(ctx context.Context, ids ...string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: append([]string{"osd", "out"}, ids...)})
}

func (c *CommandClient) OSDReweight(ctx context.Context, id string, weight float64) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: []string{"osd", "reweight", id, strconv.FormatFloat(weight, 'f', -1, 64)}})
}

func (c *CommandClient) OSDReweightByUtilization(ctx context.Context, overload *int, maxChange *float64, maxOsds *int, noIncreasing bool) (CommandResult, error) {
	args := []string{"osd", "reweight-by-utilization"}
	if overload != nil {
		args = append(args, strconv.Itoa(*overload))
	}
	if maxChange != nil {
		args = append(args, strconv.FormatFloat(*maxChange, 'f', -1, 64))
	}
	if maxOsds != nil {
		args = append(args, strconv.Itoa(*maxOsds))
	}
	args = appendBoolFlag(args, "--no-increasing", noIncreasing)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) OSDScrub(ctx context.Context, ids ...string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: append([]string{"osd", "scrub"}, ids...)})
}

func (c *CommandClient) OSDDeepScrub(ctx context.Context, ids ...string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: append([]string{"osd", "deep-scrub"}, ids...)})
}

func (c *CommandClient) OSDOkToStop(ctx context.Context, ids ...string) (map[string]any, error) {
	return c.jsonObject(ctx, append([]string{"osd", "ok-to-stop"}, ids...)...)
}

func (c *CommandClient) OSDSafeToDestroy(ctx context.Context, ids ...string) (map[string]any, error) {
	return c.jsonObject(ctx, append([]string{"osd", "safe-to-destroy"}, ids...)...)
}

func (c *CommandClient) OSDRemove(ctx context.Context, ids ...string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: append([]string{"osd", "rm"}, ids...)})
}

func (c *CommandClient) OSDDestroy(ctx context.Context, id string, yes bool) (CommandResult, error) {
	args := appendPositionals([]string{"osd", "destroy"}, id)
	args = appendBoolFlag(args, "--yes-i-really-mean-it", yes)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) OSDPurge(ctx context.Context, id string, yes bool) (CommandResult, error) {
	args := appendPositionals([]string{"osd", "purge"}, id)
	args = appendBoolFlag(args, "--yes-i-really-mean-it", yes)
	return c.Run(ctx, CommandRequest{Args: args})
}
