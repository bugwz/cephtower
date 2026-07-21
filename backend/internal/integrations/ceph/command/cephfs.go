package command

import "context"

type FSSubvolumeOptions struct {
	Size              string
	GroupName         string
	Pool              string
	UID               string
	GID               string
	Mode              string
	NamespaceIsolated bool
	Earmark           string
	Normalization     string
	CaseSensitive     bool
}

type FSSubvolumeSnapshotOptions struct {
	GroupName string
}

func (c *CommandClient) FSDump(ctx context.Context) (map[string]any, error) {
	return c.jsonObject(ctx, "fs", "dump")
}

func (c *CommandClient) FSStatus(ctx context.Context, fsName string) (map[string]any, error) {
	return c.jsonObject(ctx, appendPositionals([]string{"fs", "status"}, fsName)...)
}

func (c *CommandClient) FSList(ctx context.Context) ([]map[string]any, error) {
	var payload []map[string]any
	err := c.JSON(ctx, CommandRequest{Args: []string{"fs", "ls"}}, &payload)
	return payload, err
}

func (c *CommandClient) FSGet(ctx context.Context, fsName string) (map[string]any, error) {
	return c.jsonObject(ctx, appendPositionals([]string{"fs", "get"}, fsName)...)
}

func (c *CommandClient) FSNew(ctx context.Context, fsName string, metadataPool string, dataPool string, force bool) (CommandResult, error) {
	args := appendPositionals([]string{"fs", "new"}, fsName, metadataPool, dataPool)
	args = appendBoolFlag(args, "--force", force)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) FSRemove(ctx context.Context, fsName string, yes bool) (CommandResult, error) {
	args := appendPositionals([]string{"fs", "rm"}, fsName)
	args = appendBoolFlag(args, "--yes-i-really-mean-it", yes)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) FSFail(ctx context.Context, fsName string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: appendPositionals([]string{"fs", "fail"}, fsName)})
}

func (c *CommandClient) FSReset(ctx context.Context, fsName string, yes bool) (CommandResult, error) {
	args := appendPositionals([]string{"fs", "reset"}, fsName)
	args = appendBoolFlag(args, "--yes-i-really-mean-it", yes)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) FSSet(ctx context.Context, fsName string, variable string, value string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: appendPositionals([]string{"fs", "set"}, fsName, variable, value)})
}

func (c *CommandClient) FSSetDefault(ctx context.Context, fsName string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: appendPositionals([]string{"fs", "set-default"}, fsName)})
}

func (c *CommandClient) FSVolumeList(ctx context.Context) ([]map[string]any, error) {
	var payload []map[string]any
	err := c.JSON(ctx, CommandRequest{Args: []string{"fs", "volume", "ls"}}, &payload)
	return payload, err
}

func (c *CommandClient) FSVolumeInfo(ctx context.Context, volume string) (map[string]any, error) {
	return c.jsonObject(ctx, appendPositionals([]string{"fs", "volume", "info"}, volume)...)
}

func (c *CommandClient) FSVolumeCreate(ctx context.Context, volume string, placement string, metadataPool string, dataPool string) (CommandResult, error) {
	args := appendPositionals([]string{"fs", "volume", "create"}, volume)
	args = appendSequentialPositionals(args, placement, metadataPool, dataPool)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) FSVolumeRemove(ctx context.Context, volume string, yes bool) (CommandResult, error) {
	args := appendPositionals([]string{"fs", "volume", "rm"}, volume)
	args = appendBoolFlag(args, "--yes-i-really-mean-it", yes)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) FSSubvolumeList(ctx context.Context, volume string, groupName string) ([]map[string]any, error) {
	args := appendPositionals([]string{"fs", "subvolume", "ls"}, volume, groupName)
	var payload []map[string]any
	err := c.JSON(ctx, CommandRequest{Args: args}, &payload)
	return payload, err
}

func (c *CommandClient) FSSubvolumeInfo(ctx context.Context, volume string, subvolume string, groupName string) (map[string]any, error) {
	args := appendPositionals([]string{"fs", "subvolume", "info"}, volume, subvolume, groupName)
	return c.jsonObject(ctx, args...)
}

func (c *CommandClient) FSSubvolumeCreate(ctx context.Context, volume string, subvolume string, options FSSubvolumeOptions) (CommandResult, error) {
	args := appendPositionals([]string{"fs", "subvolume", "create"}, volume, subvolume)
	args = appendSequentialPositionals(args, options.Size, options.GroupName, options.Pool, options.UID, options.GID, options.Mode)
	args = appendBoolFlag(args, "--namespace-isolated", options.NamespaceIsolated)
	args = appendStringFlag(args, "--earmark", options.Earmark)
	args = appendStringFlag(args, "--normalization", options.Normalization)
	args = appendBoolFlag(args, "--casesensitive", options.CaseSensitive)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) FSSubvolumeRemove(ctx context.Context, volume string, subvolume string, groupName string, force bool) (CommandResult, error) {
	args := appendPositionals([]string{"fs", "subvolume", "rm"}, volume, subvolume, groupName)
	args = appendBoolFlag(args, "--force", force)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) FSSubvolumeSnapshotList(ctx context.Context, volume string, subvolume string, options FSSubvolumeSnapshotOptions) ([]map[string]any, error) {
	args := appendPositionals([]string{"fs", "subvolume", "snapshot", "ls"}, volume, subvolume, options.GroupName)
	var payload []map[string]any
	err := c.JSON(ctx, CommandRequest{Args: args}, &payload)
	return payload, err
}

func (c *CommandClient) FSSubvolumeSnapshotCreate(ctx context.Context, volume string, subvolume string, snapshot string, options FSSubvolumeSnapshotOptions) (CommandResult, error) {
	args := appendPositionals([]string{"fs", "subvolume", "snapshot", "create"}, volume, subvolume, snapshot, options.GroupName)
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) FSSubvolumeSnapshotRemove(ctx context.Context, volume string, subvolume string, snapshot string, options FSSubvolumeSnapshotOptions, force bool) (CommandResult, error) {
	args := appendPositionals([]string{"fs", "subvolume", "snapshot", "rm"}, volume, subvolume, snapshot, options.GroupName)
	args = appendBoolFlag(args, "--force", force)
	return c.Run(ctx, CommandRequest{Args: args})
}
