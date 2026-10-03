package s3

import "encoding/xml"

type BucketLifecycleRule struct {
	ID       string                  `json:"id"`
	Status   string                  `json:"status"`
	Selector BucketLifecycleSelector `json:"selector"`
	Actions  []BucketLifecycleAction `json:"actions"`
}

type BucketLifecycleSelector struct {
	Kind                  string      `json:"kind"`
	And                   bool        `json:"and"`
	Prefix                *string     `json:"prefix"`
	Tags                  []BucketTag `json:"tags"`
	ObjectSizeGreaterThan *string     `json:"object_size_greater_than"`
	ObjectSizeLessThan    *string     `json:"object_size_less_than"`
	ArchiveZone           bool        `json:"archive_zone"`
}

// Fields retain native names and decimal strings, including uint64-sized values.
type BucketLifecycleAction struct {
	Type   string            `json:"type"`
	Fields map[string]string `json:"fields"`
}

func BucketLifecycle(body []byte) ([]BucketLifecycleRule, error) {
	if err := ValidateBucketConfiguration("lifecycle", body); err != nil {
		return nil, err
	}
	var document struct {
		Rules []lifecycleField `xml:"Rule"`
	}
	if err := xml.Unmarshal(body, &document); err != nil {
		return nil, err
	}
	rules := make([]BucketLifecycleRule, 0, len(document.Rules))
	for _, source := range document.Rules {
		rule := BucketLifecycleRule{Selector: BucketLifecycleSelector{Tags: []BucketTag{}}, Actions: []BucketLifecycleAction{}}
		for _, field := range source.Children {
			switch field.XMLName.Local {
			case "ID":
				rule.ID = field.Text
			case "Status":
				rule.Status = field.Text
			case "Prefix":
				rule.Selector.Kind = "Prefix"
				rule.Selector.Prefix = &field.Text
			case "Filter":
				rule.Selector.Kind = "Filter"
				conditions := field.Children
				if len(conditions) == 1 && conditions[0].XMLName.Local == "And" {
					rule.Selector.And = true
					conditions = conditions[0].Children
				}
				for _, condition := range conditions {
					switch condition.XMLName.Local {
					case "Prefix":
						rule.Selector.Prefix = &condition.Text
					case "ObjectSizeGreaterThan":
						rule.Selector.ObjectSizeGreaterThan = &condition.Text
					case "ObjectSizeLessThan":
						rule.Selector.ObjectSizeLessThan = &condition.Text
					case "ArchiveZone":
						rule.Selector.ArchiveZone = true
					case "Tag":
						tag := BucketTag{}
						for _, part := range condition.Children {
							if part.XMLName.Local == "Key" {
								tag.Key = part.Text
							} else {
								tag.Value = part.Text
							}
						}
						rule.Selector.Tags = append(rule.Selector.Tags, tag)
					}
				}
			default:
				action := BucketLifecycleAction{Type: field.XMLName.Local, Fields: map[string]string{}}
				for _, part := range field.Children {
					action.Fields[part.XMLName.Local] = part.Text
				}
				rule.Actions = append(rule.Actions, action)
			}
		}
		rules = append(rules, rule)
	}
	return rules, nil
}
