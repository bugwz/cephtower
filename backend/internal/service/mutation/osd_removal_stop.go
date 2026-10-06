package mutation

import (
	"encoding/json"
	"strconv"
	"strings"
)

func osdRemovalStopped(id string, raw []byte) bool {
	if strings.TrimSpace(string(raw)) == "No OSD remove/replace operations reported" {
		return true
	}
	var rows []struct {
		ID *int `json:"osd_id"`
	}
	if json.Unmarshal(raw, &rows) != nil || rows == nil {
		return false
	}
	seen := map[int]bool{}
	for _, row := range rows {
		if row.ID == nil || *row.ID < 0 || *row.ID > 2147483647 || seen[*row.ID] || strconv.Itoa(*row.ID) == id {
			return false
		}
		seen[*row.ID] = true
	}
	return true
}
