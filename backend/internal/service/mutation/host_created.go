package mutation

func hostCreationConfirmed(raw []byte, parameters map[string]any) bool {
	host := optional(parameters, "hostname")
	state := "maintenance_exit"
	if boolParameter(parameters, "maintenance") {
		state = "maintenance_enter"
	}
	return hostActionStateMatches(raw, host, state) && hostUpdateMatches(raw, host, map[string]any{"labels_add": parameters["labels"]})
}
