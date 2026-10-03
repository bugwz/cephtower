package s3

import (
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
)

// BucketTag is an entry, not a map: the reference RGW stores a multimap.
type BucketTag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type tagText struct {
	Text    string     `xml:",chardata"`
	Unknown []xml.Name `xml:",any"`
}

func BucketTags(body []byte) ([]BucketTag, error) {
	if err := ValidateBucketConfiguration("tagging", body); err != nil {
		return nil, err
	}
	return parseBucketTags(body)
}

func parseBucketTags(body []byte) ([]BucketTag, error) {
	var document struct {
		Sets []struct {
			Tags []struct {
				Keys    []tagText  `xml:"Key"`
				Values  []tagText  `xml:"Value"`
				Text    string     `xml:",chardata"`
				Unknown []xml.Name `xml:",any"`
			} `xml:"Tag"`
			Text    string     `xml:",chardata"`
			Unknown []xml.Name `xml:",any"`
		} `xml:"TagSet"`
		Text    string     `xml:",chardata"`
		Unknown []xml.Name `xml:",any"`
	}
	invalid := func() ([]BucketTag, error) {
		return nil, fmt.Errorf("tagging requires one TagSet with at most 50 Tag entries, each with one Key (1-128 UTF-8 bytes) and Value (0-256 UTF-8 bytes)")
	}
	if xml.Unmarshal(body, &document) != nil || len(document.Sets) != 1 || len(document.Unknown) != 0 || strings.TrimSpace(document.Text) != "" {
		return invalid()
	}
	set := document.Sets[0]
	if len(set.Tags) > 50 || len(set.Unknown) != 0 || strings.TrimSpace(set.Text) != "" {
		return invalid()
	}
	tags := make([]BucketTag, 0, len(set.Tags))
	for _, tag := range set.Tags {
		if len(tag.Keys) != 1 || len(tag.Values) != 1 || len(tag.Unknown) != 0 || strings.TrimSpace(tag.Text) != "" {
			return invalid()
		}
		key, value := tag.Keys[0], tag.Values[0]
		if len(key.Unknown) != 0 || len(value.Unknown) != 0 || len(key.Text) == 0 || len(key.Text) > 128 || len(value.Text) > 256 {
			return invalid()
		}
		tags = append(tags, BucketTag{Key: key.Text, Value: value.Text})
	}
	// Normalize ordering only. Duplicate pairs and empty values remain significant.
	sort.Slice(tags, func(i, j int) bool {
		if tags[i].Key == tags[j].Key {
			return tags[i].Value < tags[j].Value
		}
		return tags[i].Key < tags[j].Key
	})
	return tags, nil
}
