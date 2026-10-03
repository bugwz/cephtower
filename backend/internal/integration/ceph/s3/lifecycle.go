package s3

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type lifecycleField struct {
	XMLName  xml.Name
	Text     string           `xml:",chardata"`
	Children []lifecycleField `xml:",any"`
}

// Validate the native rule structure and supported local semantic checks.
// RGW remains authoritative for placement and cross-rule constraints.
func validateBucketLifecycle(body []byte) error {
	var document struct {
		Rules []struct {
			Fields []lifecycleField `xml:",any"`
			Text   string           `xml:",chardata"`
		} `xml:"Rule"`
		Unknown []xml.Name `xml:",any"`
		Text    string     `xml:",chardata"`
	}
	invalid := fmt.Errorf("lifecycle requires Rule entries with one Enabled/Disabled Status, a Filter or Prefix, and at least one expiration or transition action")
	if xml.Unmarshal(body, &document) != nil || len(document.Rules) == 0 || len(document.Unknown) > 0 || strings.TrimSpace(document.Text) != "" {
		return invalid
	}
	for _, rule := range document.Rules {
		if strings.TrimSpace(rule.Text) != "" {
			return invalid
		}
		counts := map[string]int{}
		actions := 0
		usingDays, usingDate := false, false
		classes := map[string]map[string]bool{"Transition": {}, "NoncurrentVersionTransition": {}}
		for _, field := range rule.Fields {
			name := field.XMLName.Local
			counts[name]++
			switch name {
			case "ID", "Prefix", "Status":
				if len(field.Children) > 0 {
					return invalid
				}
				if name == "Status" && field.Text != "Enabled" && field.Text != "Disabled" {
					return invalid
				}
				if name == "ID" && len(field.Text) > 255 {
					return fmt.Errorf("lifecycle rule ID exceeds 255 UTF-8 bytes")
				}
			case "Filter":
				if strings.TrimSpace(field.Text) != "" {
					return invalid
				}
				if err := validateLifecycleFilter(field); err != nil {
					return err
				}
			case "Expiration", "NoncurrentVersionExpiration", "AbortIncompleteMultipartUpload", "Transition", "NoncurrentVersionTransition":
				if strings.TrimSpace(field.Text) != "" || len(field.Children) == 0 {
					return invalid
				}
				if err := validateLifecycleAction(field); err != nil {
					return err
				}
				effective := false
				for _, child := range field.Children {
					key := child.XMLName.Local
					if name == "Expiration" || name == "Transition" {
						usingDays = usingDays || key == "Days"
						usingDate = usingDate || key == "Date"
					}
					if key == "StorageClass" {
						if classes[name][child.Text] {
							return fmt.Errorf("duplicate lifecycle %s storage class", name)
						}
						classes[name][child.Text] = true
					}
					effective = effective || key != "ExpiredObjectDeleteMarker" || child.Text == "true"
				}
				if effective {
					actions++
				}
			default:
				return invalid
			}
			if name != "Transition" && name != "NoncurrentVersionTransition" && counts[name] > 1 {
				return invalid
			}
		}
		if counts["Status"] != 1 || counts["Filter"]+counts["Prefix"] != 1 || actions == 0 {
			return invalid
		}
		if usingDays && usingDate {
			return fmt.Errorf("lifecycle current-version expiration and transitions cannot mix Days and Date")
		}
	}
	return nil
}

func validateLifecycleFilter(filter lifecycleField) error {
	invalid := fmt.Errorf("invalid lifecycle Filter: use scalar conditions and Tag entries, optionally inside one exclusive And element")
	children := filter.Children
	for _, child := range children {
		if child.XMLName.Local == "And" {
			if len(children) != 1 || strings.TrimSpace(child.Text) != "" {
				return invalid
			}
			children = child.Children
			break
		}
	}
	counts := map[string]int{}
	sizes := map[string]string{}
	for _, child := range children {
		name := child.XMLName.Local
		counts[name]++
		switch name {
		case "Prefix", "ObjectSizeGreaterThan", "ObjectSizeLessThan", "ArchiveZone":
			if len(child.Children) > 0 || counts[name] > 1 {
				return invalid
			}
			if name == "ArchiveZone" && strings.TrimSpace(child.Text) != "" {
				return invalid
			}
			if name == "ObjectSizeGreaterThan" || name == "ObjectSizeLessThan" {
				if child.Text != "" {
					for _, digit := range child.Text {
						if digit < '0' || digit > '9' {
							return fmt.Errorf("lifecycle object sizes must be unsigned decimal byte counts")
						}
					}
					if _, err := strconv.ParseUint(child.Text, 10, 64); err != nil {
						return fmt.Errorf("lifecycle object size exceeds uint64 range")
					}
					sizes[name] = child.Text
				}
			}
		case "Tag":
			if strings.TrimSpace(child.Text) != "" {
				return invalid
			}
			fields := map[string]bool{}
			for _, field := range child.Children {
				key := field.XMLName.Local
				if (key != "Key" && key != "Value") || fields[key] || len(field.Children) > 0 {
					return invalid
				}
				fields[key] = true
			}
		default:
			return invalid
		}
	}
	greater, less := sizes["ObjectSizeGreaterThan"], sizes["ObjectSizeLessThan"]
	if greater != "" && less != "" {
		minimum, _ := strconv.ParseUint(greater, 10, 64)
		maximum, _ := strconv.ParseUint(less, 10, 64)
		if maximum <= minimum {
			return fmt.Errorf("lifecycle maximum object size must exceed the minimum")
		}
		// The bundled RGW compares its stored string members in decode_xml.
		// Preserve the submitted text and report that limitation, never silently
		// pad/rewrite bounds to bypass native validation.
		if less <= greater {
			return fmt.Errorf("reference RGW rejects these object size bounds by textual comparison; verify target Ceph support before submitting this range")
		}
	}
	return nil
}

func validateLifecycleAction(action lifecycleField) error {
	name := action.XMLName.Local
	allowed := map[string][]string{
		"Expiration":                     {"Days", "Date", "ExpiredObjectDeleteMarker"},
		"NoncurrentVersionExpiration":    {"NoncurrentDays", "NewerNoncurrentVersions"},
		"AbortIncompleteMultipartUpload": {"DaysAfterInitiation"},
		"Transition":                     {"Days", "Date", "StorageClass"},
		"NoncurrentVersionTransition":    {"NoncurrentDays", "StorageClass"},
	}[name]
	counts := map[string]int{}
	invalid := fmt.Errorf("invalid lifecycle %s fields: check required, exclusive and duplicate scalar elements", name)
	for _, child := range action.Children {
		key := child.XMLName.Local
		known := false
		for _, field := range allowed {
			if key == field {
				known = true
				break
			}
		}
		if !known || len(child.Children) > 0 {
			return invalid
		}
		counts[key]++
		if counts[key] > 1 {
			return invalid
		}
		if key == "Days" || key == "NoncurrentDays" || key == "DaysAfterInitiation" || key == "NewerNoncurrentVersions" {
			minimum := uint64(1)
			if name == "Transition" || name == "NoncurrentVersionTransition" || key == "NewerNoncurrentVersions" {
				minimum = 0
			}
			if child.Text == "" {
				return fmt.Errorf("lifecycle %s must contain an integer", key)
			}
			for _, digit := range child.Text {
				if digit < '0' || digit > '9' {
					return fmt.Errorf("lifecycle %s must contain decimal digits only", key)
				}
			}
			number, err := strconv.ParseUint(child.Text, 10, 31)
			if err != nil || number < minimum {
				return fmt.Errorf("lifecycle %s must be between %d and 2147483647", key, minimum)
			}
		}
		if key == "ExpiredObjectDeleteMarker" && child.Text != "true" && child.Text != "false" {
			return fmt.Errorf("ExpiredObjectDeleteMarker must be exactly true or false")
		}
		if key == "Date" && !validLifecycleDate(child.Text) {
			return fmt.Errorf("lifecycle Date must be a valid UTC midnight within the native clock range")
		}
	}
	switch name {
	case "Expiration":
		if counts["Days"]+counts["Date"]+counts["ExpiredObjectDeleteMarker"] != 1 {
			return invalid
		}
	case "Transition":
		if counts["Days"]+counts["Date"] != 1 || counts["StorageClass"] != 1 {
			return invalid
		}
	case "NoncurrentVersionTransition":
		if counts["NoncurrentDays"] != 1 || counts["StorageClass"] != 1 {
			return invalid
		}
	case "NoncurrentVersionExpiration":
		if counts["NoncurrentDays"] != 1 {
			return invalid
		}
	case "AbortIncompleteMultipartUpload":
		if counts["DaysAfterInitiation"] != 1 {
			return invalid
		}
	}
	return nil
}

// Ceph accepts partial ISO dates and UTC times. Reject whitespace termination
// and calendar normalization instead of silently changing the user's intent.
var lifecycleDatePattern = regexp.MustCompile(`^[0-9]{4}(-[0-9]{2}(-[0-9]{2}(T[0-9]{2}(:[0-9]{2}(:[0-9]{2}(\.[0-9]{1,9})?)?)?Z)?)?)?$`)

func validLifecycleDate(value string) bool {
	if !lifecycleDatePattern.MatchString(value) {
		return false
	}
	for _, layout := range []string{"2006", "2006-01", "2006-01-02", "2006-01-02T15Z", "2006-01-02T15:04Z", time.RFC3339Nano} {
		parsed, err := time.Parse(layout, value)
		if err != nil {
			continue
		}
		seconds := parsed.Unix()
		// real_clock uses unsigned 64-bit nanoseconds; reject wraparound.
		return seconds >= 0 && uint64(seconds) <= ^uint64(0)/1_000_000_000 &&
			seconds%86400 == 0 && parsed.Nanosecond() == 0
	}
	return false
}
