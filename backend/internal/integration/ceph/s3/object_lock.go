package s3

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

type BucketObjectLockConfiguration struct {
	Enabled          bool                    `json:"enabled"`
	DefaultRetention *BucketDefaultRetention `json:"default_retention"`
}
type BucketDefaultRetention struct {
	Mode  string  `json:"mode"`
	Days  *string `json:"days"`
	Years *string `json:"years"`
}

// Preserve native values on reads; writes additionally validate retention semantics.
func BucketObjectLock(body []byte) (BucketObjectLockConfiguration, error) {
	result := BucketObjectLockConfiguration{}
	if err := validateConfigurationXML(body, "ObjectLockConfiguration"); err != nil {
		return result, err
	}
	var root lifecycleField
	if err := xml.Unmarshal(body, &root); err != nil {
		return result, err
	}
	invalid := fmt.Errorf("invalid native ObjectLockConfiguration structure")
	if strings.TrimSpace(root.Text) != "" {
		return result, invalid
	}
	seen := map[string]bool{}
	for _, child := range root.Children {
		name := child.XMLName.Local
		if seen[name] {
			return result, invalid
		}
		seen[name] = true
		switch name {
		case "ObjectLockEnabled":
			if child.Text != "Enabled" || len(child.Children) > 0 {
				return result, invalid
			}
			result.Enabled = true
		case "Rule":
			if strings.TrimSpace(child.Text) != "" || len(child.Children) != 1 || child.Children[0].XMLName.Local != "DefaultRetention" {
				return result, invalid
			}
			retention := child.Children[0]
			if strings.TrimSpace(retention.Text) != "" {
				return result, invalid
			}
			fields := map[string]string{}
			for _, field := range retention.Children {
				key := field.XMLName.Local
				if (key != "Mode" && key != "Days" && key != "Years") || len(field.Children) > 0 {
					return result, invalid
				}
				if _, duplicate := fields[key]; duplicate {
					return result, invalid
				}
				fields[key] = field.Text
			}
			days, hasDays := fields["Days"]
			years, hasYears := fields["Years"]
			if fields["Mode"] == "" || hasDays == hasYears {
				return result, invalid
			}
			result.DefaultRetention = &BucketDefaultRetention{Mode: fields["Mode"]}
			if hasDays {
				result.DefaultRetention.Days = &days
			} else {
				result.DefaultRetention.Years = &years
			}
		default:
			return result, invalid
		}
	}
	if !result.Enabled {
		return result, invalid
	}
	return result, nil
}

func ValidatedBucketObjectLock(body []byte) (BucketObjectLockConfiguration, error) {
	configuration, err := BucketObjectLock(body)
	if err != nil || configuration.DefaultRetention == nil {
		return configuration, err
	}
	retention := configuration.DefaultRetention
	if retention.Mode != "GOVERNANCE" && retention.Mode != "COMPLIANCE" {
		return configuration, fmt.Errorf("object lock mode must be GOVERNANCE or COMPLIANCE")
	}
	period := retention.Days
	if period == nil {
		period = retention.Years
	}
	for _, digit := range *period {
		if digit < '0' || digit > '9' {
			return configuration, fmt.Errorf("object lock period must contain decimal digits only")
		}
	}
	number, err := strconv.ParseUint(*period, 10, 31)
	if err != nil || number == 0 {
		return configuration, fmt.Errorf("object lock period must be between 1 and 2147483647")
	}
	*period = strconv.FormatUint(number, 10)
	return configuration, nil
}
