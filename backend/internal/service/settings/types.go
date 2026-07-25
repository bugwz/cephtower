package settings

import "errors"

var (
	ErrGroupNotFound = errors.New("ceph setting group not found")
	ErrInvalidGroup  = errors.New("setting does not belong to group")
	ErrInvalidReply  = errors.New("unexpected ceph settings response")
)

type Setting struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Default   bool   `json:"default"`
	Sensitive bool   `json:"sensitive"`
	ValueSet  bool   `json:"value_set"`
	Value     any    `json:"value"`
}

type Group struct {
	Name     string    `json:"name"`
	Settings []Setting `json:"settings"`
}
