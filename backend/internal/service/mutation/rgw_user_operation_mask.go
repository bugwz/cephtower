package mutation

import (
	"encoding/json"
	"strings"
)

func rgwUserOperationMaskMatches(raw []byte, expected, uid string) bool {
	var info struct {
		UID  *string `json:"full_user_id"`
		Mask *string `json:"op_mask"`
	}
	return uid != "" && expected != "" && json.Unmarshal(raw, &info) == nil &&
		info.UID != nil && *info.UID == uid && info.Mask != nil &&
		*info.Mask == strings.ReplaceAll(expected, ",", ", ")
}
