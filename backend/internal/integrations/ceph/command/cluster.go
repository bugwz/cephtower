package command

import (
	"context"
	"encoding/json"
)

func (c *CommandClient) Status(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "status")
}

func (c *CommandClient) Health(ctx context.Context, detail bool) (map[string]any, error) {
	args := []string{"health"}
	if detail {
		args = append(args, "detail")
	}
	return c.jsonObject(ctx, args...)
}

func (c *CommandClient) DF(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "df")
}

func (c *CommandClient) Features(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "features")
}

func (c *CommandClient) FSID(ctx context.Context) (string, error) {
	return c.Text(ctx, "fsid")
}

func (c *CommandClient) QuorumStatus(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "quorum_status")
}

func (c *CommandClient) Report(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "report")
}

func (c *CommandClient) NodeList(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "node", "ls")
}

func (c *CommandClient) TimeSyncStatus(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "time-sync-status")
}

func (c *CommandClient) Versions(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "versions")
}

func (c *CommandClient) LogLast(ctx context.Context, num *int, channel string) (map[string]any, error) {
	args := []string{"log", "last"}
	args = appendIntFlag(args, "--num", num)
	args = appendStringFlag(args, "--channel", channel)
	return c.jsonObject(ctx, args...)
}

func (c *CommandClient) Log(ctx context.Context, message string, level string, channel string) (CommandResult, error) {
	args := []string{"log"}
	args = appendPositionals(args, message)
	args = appendStringFlag(args, "--level", level)
	args = appendStringFlag(args, "--channel", channel)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) HealthMute(ctx context.Context, code string, sticky bool, ttl string) (CommandResult, error) {
	args := []string{"health", "mute"}
	args = appendPositionals(args, code)
	args = appendBoolFlag(args, "--sticky", sticky)
	args = appendStringFlag(args, "--ttl", ttl)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) HealthUnmute(ctx context.Context, code string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: appendPositionals([]string{"health", "unmute"}, code)})
}

func (c *CommandClient) AuthList(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "auth", "ls")
}

func (c *CommandClient) AuthGet(ctx context.Context, entity string) (string, error) {
	return c.Text(ctx, "auth", "get", entity)
}

func (c *CommandClient) AuthGetKey(ctx context.Context, entity string) (string, error) {
	return c.Text(ctx, "auth", "get-key", entity)
}

func (c *CommandClient) AuthPrintKey(ctx context.Context, entity string) (string, error) {
	return c.Text(ctx, "auth", "print-key", entity)
}

func (c *CommandClient) AuthAdd(ctx context.Context, entity string, caps ...string) (CommandResult, error) {
	args := appendPositionals([]string{"auth", "add"}, entity)
	args = append(args, caps...)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) AuthCaps(ctx context.Context, entity string, caps ...string) (CommandResult, error) {
	args := appendPositionals([]string{"auth", "caps"}, entity)
	args = append(args, caps...)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) AuthGetOrCreate(ctx context.Context, entity string, caps ...string) (string, error) {
	args := appendPositionals([]string{"auth", "get-or-create"}, entity)
	args = append(args, caps...)
	return c.Text(ctx, args...)
}

func (c *CommandClient) AuthGetOrCreateKey(ctx context.Context, entity string, caps ...string) (string, error) {
	args := appendPositionals([]string{"auth", "get-or-create-key"}, entity)
	args = append(args, caps...)
	return c.Text(ctx, args...)
}

func (c *CommandClient) AuthRemove(ctx context.Context, entity string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: appendPositionals([]string{"auth", "rm"}, entity)})
}

func (c *CommandClient) AuthImport(ctx context.Context, keyring []byte) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: []string{"auth", "import", "-i", "-"}, Input: keyring})
}

func (c *CommandClient) ConfigDump(ctx context.Context) ([]map[string]any, error) {
	var payload []map[string]any
	err := c.JSON(ctx, CommandRequest{Args: []string{"config", "dump"}}, &payload)
	return payload, err
}

func (c *CommandClient) ConfigList(ctx context.Context) ([]map[string]any, error) {
	var payload []map[string]any
	err := c.JSON(ctx, CommandRequest{Args: []string{"config", "ls"}}, &payload)
	return payload, err
}

func (c *CommandClient) ConfigGet(ctx context.Context, who string, option string) (string, error) {
	return c.Text(ctx, "config", "get", who, option)
}

func (c *CommandClient) ConfigSet(ctx context.Context, who string, option string, value string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: appendPositionals([]string{"config", "set"}, who, option, value)})
}

func (c *CommandClient) ConfigRemove(ctx context.Context, who string, option string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: appendPositionals([]string{"config", "rm"}, who, option)})
}

func (c *CommandClient) ConfigShow(ctx context.Context, who string) (map[string]any, error) {
	return c.jsonObject(ctx, "config", "show", who)
}

func (c *CommandClient) ConfigShowWithDefaults(ctx context.Context, who string) (map[string]any, error) {
	return c.jsonObject(ctx, "config", "show-with-defaults", who)
}

func (c *CommandClient) ConfigKeyDump(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "config-key", "dump")
}

func (c *CommandClient) ConfigKeyList(ctx context.Context) ([]string, error) {
	var payload []string
	err := c.JSON(ctx, CommandRequest{Args: []string{"config-key", "ls"}}, &payload)
	return payload, err
}

func (c *CommandClient) ConfigKeyGet(ctx context.Context, key string) (string, error) {
	return c.Text(ctx, "config-key", "get", key)
}

func (c *CommandClient) ConfigKeyExists(ctx context.Context, key string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: appendPositionals([]string{"config-key", "exists"}, key)})
}

func (c *CommandClient) ConfigKeySet(ctx context.Context, key string, value []byte) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: []string{"config-key", "set", key, "-i", "-"}, Input: value})
}

func (c *CommandClient) ConfigKeyRemove(ctx context.Context, key string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: appendPositionals([]string{"config-key", "rm"}, key)})
}

func (c *CommandClient) jsonObject(ctx context.Context, args ...string) (map[string]any, error) {
	var payload map[string]any
	err := c.JSON(ctx, CommandRequest{Args: args}, &payload)
	return payload, err
}

func (c *CommandClient) rawJSON(ctx context.Context, args ...string) (json.RawMessage, error) {
	return c.RawJSON(ctx, args...)
}
