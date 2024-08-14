package apigen

import (
	"fmt"
	"strings"
)

type ruleSet int

const (
	requiredRule ruleSet = 1 << iota
	defaultRule
	enumRule
	minRule
	maxRule
	greaterRule
	lessRule
)

type validator struct {
	rules      ruleSet
	paramName  string
	defaultVal string
	enum       []string
	min        string
	max        string
	greater    string
	less       string
}

func parseValidator(s string) (*validator, error) {
	var v validator
	var err error

	for _, entry := range strings.Split(s, ",") {
		switch {
		case strings.HasPrefix(entry, "paramname="):
			v.paramName = strings.TrimPrefix(entry, "paramname=")

		case entry == "required":
			v.rules |= requiredRule

		case strings.HasPrefix(entry, "default="):
			v.rules |= defaultRule
			v.defaultVal = strings.TrimPrefix(entry, "default=")

		case strings.HasPrefix(entry, "enum="):
			v.rules |= enumRule
			v.enum = strings.Split(strings.TrimPrefix(entry, "enum="), "|")

		case strings.HasPrefix(entry, "min="):
			v.rules |= minRule
			v.min = strings.TrimPrefix(entry, "min=")

		case strings.HasPrefix(entry, "max="):
			v.rules |= maxRule
			v.max = strings.TrimPrefix(entry, "max=")

		case strings.HasPrefix(entry, ">="):
			v.rules |= minRule
			v.min = strings.TrimPrefix(entry, ">=")

		case strings.HasPrefix(entry, "<="):
			v.rules |= maxRule
			v.max = strings.TrimPrefix(entry, "<=")

		case strings.HasPrefix(entry, ">"):
			v.rules |= greaterRule
			v.greater = strings.TrimPrefix(entry, ">")

		case strings.HasPrefix(entry, "<"):
			v.rules |= lessRule
			v.less = strings.TrimPrefix(entry, "<")
			
		default:
			err = fmt.Errorf("%s: unknown rule", entry)
		}

		if err != nil {
			return nil, err
		}
	}

	return &v, nil
}
