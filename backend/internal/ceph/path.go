package ceph

import "regexp"

var pathTemplateParamRE = regexp.MustCompile(`\{[^}/]+\}`)
