package mutation

import (
	"bytes"
	"encoding/json"
	"io"
	"regexp"
	"strings"
)

type nfsClientRule struct {
	Addresses  []string `json:"addresses"`
	AccessType string   `json:"access_type"`
	Squash     string   `json:"squash"`
}

var nfsClientAddress = regexp.MustCompile(`^[A-Za-z0-9_.:@*?/-]+$`)

func nfsClients(value any) ([]nfsClientRule, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, invalid("invalid NFS clients")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var clients []nfsClientRule
	if decoder.Decode(&clients) != nil || clients == nil {
		return nil, invalid("clients must be an array of client rules")
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return nil, invalid("invalid NFS clients")
	}
	for _, client := range clients {
		if len(client.Addresses) == 0 {
			return nil, invalid("client addresses cannot be empty")
		}
		for _, address := range client.Addresses {
			if !nfsClientAddress.MatchString(address) || len(address) > 255 {
				return nil, invalid("invalid NFS client address")
			}
		}
		if client.AccessType != "" && client.AccessType != "RO" && client.AccessType != "RW" && client.AccessType != "NONE" {
			return nil, invalid("invalid NFS client access_type")
		}
		switch strings.ToLower(client.Squash) {
		case "", "root", "root_squash", "rootsquash", "rootid", "root_id_squash", "rootidsquash", "all", "all_squash", "allsquash", "all_anomnymous", "allanonymous", "no_root_squash", "none", "noidsquash":
		default:
			return nil, invalid("invalid NFS client squash")
		}
	}
	return clients, nil
}

func nfsClientsMatch(actual, expected any) bool {
	left, err := nfsClients(actual)
	right, otherErr := nfsClients(expected)
	if err != nil || otherErr != nil {
		return false
	}
	a, _ := json.Marshal(left)
	b, _ := json.Marshal(right)
	return bytes.Equal(a, b)
}
