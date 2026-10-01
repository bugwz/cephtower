package mutation

func upgradeTargetArgs(parameters map[string]any) ([]string, error) {
	version, image := optional(parameters, "version"), optional(parameters, "image")
	if (version == "") == (image == "") {
		return nil, invalid("exactly one upgrade version or image is required")
	}
	if image != "" {
		value, err := required(parameters, "image")
		if err != nil {
			return nil, err
		}
		return []string{"--image", value}, nil
	}
	value, err := required(parameters, "version")
	if err != nil {
		return nil, err
	}
	return []string{"--ceph-version", value}, nil
}
