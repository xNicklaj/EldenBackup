package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
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

func Test_BackupFile(t *testing.T) {
	ViperSetup()
	fp := viper.GetString("savefiledirectory")
	fp = strings.Replace(fp, "SteamID", GetSteamID(), -1)
	fp_out := viper.GetString("backupdirectory")

	// Create save file directory if not exists
	if _, err := os.Stat(ResolvePath(fp)); err != nil {
		err = os.MkdirAll(ResolvePath(fp), 0660)
		assert.Nil(t, err)
	}

	// Create save file if not exists
	if _, err := os.Stat(ResolvePath(fp + GetSaveName())); err != nil {
		_, err = os.Create(ResolvePath(fp + GetSaveName()))
		assert.Nil(t, err)
	}

	// Create backup directory if not exists
	if _, err := os.Stat(ResolvePath(fp_out)); err != nil {
		err = os.MkdirAll(ResolvePath(fp_out), 0660)
		assert.Nil(t, err)
	}

	bck_path := BackupFile(ResolvePath(fp)+GetSaveName(), BCK_MANUAL)
	assert.FileExists(t, bck_path)
}
