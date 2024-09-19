package main

import (
	"regexp"
)

func fix_star_id(value string) string {
	annoying_star_prefix := regexp.MustCompile(`\d+-(\d+)`)
	match := annoying_star_prefix.FindStringSubmatch(value)
	if match != nil {
		return match[1]
	} else {
		return value
	}
}
