package dashboard

import "regexp"

var pathTemplateParamRE = regexp.MustCompile(`\{[^}/]+\}`)
