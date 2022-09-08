package main

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ResolveAppdata(t *testing.T) {
	var test_appdata = "%appdata%\\EldenRing"
	var isAbs = filepath.IsAbs(ResolvePath(test_appdata))
	assert.Equal(t, isAbs, true, "Cannot resolve %appdata%.")
}

func Test_ResolveLocalappdata(t *testing.T) {
	var test_localappdata = "%localappdata%\\EldenRing"
	var isAbs = filepath.IsAbs(ResolvePath(test_localappdata))
	assert.Equal(t, isAbs, true, "Cannot resolve %localappdata%.")
}

func Test_ResolveUserprofile(t *testing.T) {
	var test_userprofile = "%userprofile%\\EldenRing"
	var isAbs = filepath.IsAbs(ResolvePath(test_userprofile))
	assert.Equal(t, isAbs, true, "Cannot resolve %userprofile%.")
}
