package command

import "strconv"

func appendStringFlag(args []string, name string, value string) []string {
	if value == "" {
		return args
	}
	return append(args, name, value)
}

func appendIntFlag(args []string, name string, value *int) []string {
	if value == nil {
		return args
	}
	return append(args, name, strconv.Itoa(*value))
}

func appendFloatFlag(args []string, name string, value *float64) []string {
	if value == nil {
		return args
	}
	return append(args, name, strconv.FormatFloat(*value, 'f', -1, 64))
}

func appendBoolFlag(args []string, name string, enabled bool) []string {
	if !enabled {
		return args
	}
	return append(args, name)
}

func appendRepeatedFlag(args []string, name string, values []string) []string {
	for _, value := range values {
		if value != "" {
			args = append(args, name, value)
		}
	}
	return args
}

func appendPositionals(args []string, values ...string) []string {
	for _, value := range values {
		if value != "" {
			args = append(args, value)
		}
	}
	return args
}

func appendSequentialPositionals(args []string, values ...string) []string {
	for _, value := range values {
		if value == "" {
			return args
		}
		args = append(args, value)
	}
	return args
}
