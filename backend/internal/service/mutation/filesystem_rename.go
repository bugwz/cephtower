package mutation

import "encoding/json"

// Verify names from the native fs volume ls array, not from command acceptance.
func filesystemRenameMatches(previous, next string, data []byte) bool {
	var volumes []struct {
		Name string `json:"name"`
	}
	if json.Unmarshal(data, &volumes) != nil || previous == "" || next == "" || previous == next {
		return false
	}
	found := false
	seen := map[string]bool{}
	for _, volume := range volumes {
		if volume.Name == "" || volume.Name == previous || seen[volume.Name] {
			return false
		}
		seen[volume.Name] = true
		if volume.Name == next {
			found = true
		}
	}
	return found
}
