package main

import "path/filepath"

type FilenameSet []string

func (fs FilenameSet) Match(fn string) bool {
	if fs == nil {
		return true
	}
	for _, matcher := range fs {
		if matched, _ := filepath.Match(matcher, fn); matched {
			return true
		}
	}
	return false
}
