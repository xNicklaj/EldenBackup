package main

//go:generate goversioninfo -icon=./Icon.ico

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/ChromeTemp/Popup"
	"github.com/abdfnx/gosh"
	"github.com/fsnotify/fsnotify"
	"github.com/getlantern/systray"
	"github.com/gouniverse/utils"
	"github.com/mitchellh/go-ps"
	"github.com/spf13/viper"
)

var wHandle *fsnotify.Watcher
var quit chan bool
var STEAMID string
var Logger *log.Logger

const APP_TITLE = "Elden Backup"

// const SAVE_PATH = "%appdata%\\EldenRing\\"
var SAVE_PATH string

const (
	BCK_STARTUP int = 0
	BCK_MANUAL  int = 1
	BCK_AUTO    int = 2
	BCK_TIMEOUT int = 3
)

const (
	LOG_FATAL int = 0
	LOG_DEBUG int = 1
)

func GetSteamID() string {
	if viper.GetInt("SteamID") > 0 {
		return strconv.Itoa(viper.GetInt("SteamID"))
	}
	err, out, _ := gosh.RunOutput(`$SteamID = reg query HKEY_CURRENT_USER\SOFTWARE\Valve\Steam\ActiveProcess /v ActiveUser; $__SteamID = [uint32]($SteamID[2] -replace ".*(?=0x)",""); echo $__SteamID`)
	check(err, true)

	steamid := strings.Replace(out, "\n", "", -1)
	fs, err := os.ReadDir(ResolvePath("%appdata%\\EldenRing\\"))
	if !check(err, false) {
		return "0"
	}

	for _, f := range fs {
		if f.Name() == steamid {
			return steamid
		}
	}

	for _, f := range fs {
		if _, err := strconv.Atoi(f.Name()); err == nil {
			return f.Name()
		}
	}

	return "0"
}

func LimitSaveFiles(files []os.DirEntry, limit int) {
	if len(files) > limit {
		for i := limit; i < len(files); i++ {
			err := os.Remove(ResolvePath(viper.GetString("backupdirectory")) + files[i].Name())
			if !check(err, false) {
				if viper.GetBool("EnableLogging") {
					Log(viper.GetString("LogsPath"), "Could not delete file "+files[i].Name()+". Check that your current user has full permissions over that file.", LOG_DEBUG)
				}
			}
		}
	}
}

func ListBackupsOfType(BCK_TYPE int) []os.DirEntry {
	var filtered []os.DirEntry
	fs, err := os.ReadDir(ResolvePath(viper.GetString("backupdirectory")))
	check(err, true)
	for _, f := range fs {

		switch BCK_TYPE {
		case BCK_TIMEOUT:
			if strings.HasSuffix(strings.Trim(f.Name(), filepath.Ext(f.Name())), "T") {
				filtered = append(filtered, f)
			}
		case BCK_AUTO:
			if strings.HasSuffix(strings.Trim(f.Name(), filepath.Ext(f.Name())), "A") {
				filtered = append(filtered, f)
			}
		case BCK_MANUAL:
			if strings.HasSuffix(strings.Trim(f.Name(), filepath.Ext(f.Name())), "M") {
				filtered = append(filtered, f)
			}
		case BCK_STARTUP:
			if strings.HasSuffix(strings.Trim(f.Name(), filepath.Ext(f.Name())), "S") {
				filtered = append(filtered, f)
			}
		}
	}
	return filtered
}

func GetSaveName() string {
	if viper.GetBool("UseSeamlessCoop") {
		return "ER0000.co2"
	}
	return "ER0000.sl2"
}

func CopyFile(src string, dst string) {
	srcFile, err := os.Open(src)
	readError := check(err, false)
	if !readError {
		return
	}
	defer srcFile.Close()

	destFile, err := os.Create(dst) // creates if file doesn't exist
	if check(err, false) {
		defer destFile.Close()

		_, err = io.Copy(destFile, srcFile) // check first var for number of bytes copied
		if check(err, false) {
			err = destFile.Sync()
			check(err, false)
		}
	}

}

func check(err error, exit bool) bool {
	if err != nil {
		if viper.GetBool("EnableLogging") {
			Log(viper.GetString("LogsPath"), "Error "+err.Error()+" encountered.", LOG_FATAL)
			Log(viper.GetString("LogsPath"), string(debug.Stack()), LOG_FATAL)
		}
		if exit {
			Popup.Alert("Elden Backup", "Error: "+err.Error())
			if viper.GetBool("EnableLogging") {
				Log(viper.GetString("LogsPath"), "Shutting down the application due to an error.", LOG_FATAL)
			}
			os.Exit(-1)
		}
		return false
	}
	return true
}

func ResolvePath(path string) string {
	s := strings.Replace(path, "%appdata%", os.Getenv("APPDATA"), -1)
	s = strings.Replace(s, "%APPDATA%", os.Getenv("APPDATA"), -1)
	s = strings.Replace(s, "%localappdata%", os.Getenv("LOCALAPPDATA"), -1)
	s = strings.Replace(s, "%LOCALAPPDATA%", os.Getenv("LOCALAPPDATA"), -1)
	s = strings.Replace(s, "%userprofile%", os.Getenv("USERPROFILE"), -1)
	s = strings.Replace(s, "%USERPROFILE%", os.Getenv("USERPROFILE"), -1)
	return s
}

func BackupFile(inp_path string, mode int) string {
	bck_path := ResolvePath(viper.GetString("BackupDirectory"))
	file_path := ResolvePath(inp_path)

	ext := filepath.Ext(file_path)
	filename := filepath.Base(file_path)
	filename = strings.Trim(filename, ext)

	ctime := time.Now()

	switch mode {
	case BCK_STARTUP:
		bck_path = bck_path + filename + "-" + STEAMID + "-" + ctime.Format("20060102_1504") + "S" + ext
	case BCK_AUTO:
		bck_path = bck_path + filename + "-" + STEAMID + "-" + ctime.Format("20060102") + "A" + ext
	case BCK_MANUAL:
		bck_path = bck_path + filename + "-" + STEAMID + "-" + ctime.Format("20060102_1504") + "M" + ext
	case BCK_TIMEOUT:
		bck_path = bck_path + filename + "-" + STEAMID + "-" + ctime.Format("20060102_1504") + "T" + ext
	}
	CopyFile(file_path, bck_path)
	if viper.GetBool("EnableLogging") {
		Log(viper.GetString("LogsPath"), "Backup executed at "+bck_path+".", LOG_DEBUG)
	}

	current_files := utils.ArrayReverse(ListBackupsOfType(mode))
	if mode == BCK_TIMEOUT && viper.GetInt("limittimeoutbackups") > 0 {
		LimitSaveFiles(current_files, viper.GetInt("limittimeoutbackups"))
	}
	if mode == BCK_AUTO && viper.GetInt("limitautobackups") > 0 {
		LimitSaveFiles(current_files, viper.GetInt("limitautobackups"))
	}
	return bck_path
}

func StartWatcher(w *fsnotify.Watcher) {
	go func() {
		for {
			select {
			case event, ok := <-w.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write != fsnotify.Write {
					break
				}
				if filepath.Base(event.Name) == GetSaveName() {
					BackupFile(event.Name, BCK_AUTO)
				}
			case err, ok := <-w.Errors:
				if !ok {
					return
				}
				check(err, false)
			}
		}
	}()
	if err := w.Add(ResolvePath(SAVE_PATH)); err != nil {
		check(err, false)
	}
}

func IntervalledBackup(delay int) {
	ticker := time.NewTicker(time.Duration(delay) * time.Minute)
	for {
		select {
		case <-ticker.C:
			BackupFile(SAVE_PATH+GetSaveName(), BCK_TIMEOUT)
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func Log(path string, s string, level int) {
	f, err := os.OpenFile(ResolvePath(path), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	logger := log.New(f, "", log.LstdFlags)
	switch level {
	case LOG_DEBUG:
		log.Println(s)
		logger.Println(s)
	case LOG_FATAL:
		logger.Fatal(s)
		log.Fatal(s)
	}

}

func ViperSetup() error {
	var err error = nil
	viper.SetDefault("BackupDirectory", "%appdata%\\EldenRingBackup\\")
	viper.SetDefault("BackupOnStartup", true)
	viper.SetDefault("BackupIntervalTimeout", 10)
	viper.SetDefault("UseSeamlessCoop", true)
	viper.SetDefault("LimitTimeoutBackups", 0)
	viper.SetDefault("LimitAutoBackups", 0)
	viper.SetDefault("SavefileDirectory", "%appdata%\\EldenRing\\SteamID\\")
	viper.SetDefault("EnableLogging", false)
	viper.SetDefault("EnableSaveListener", true)
	viper.SetDefault("LogsPath", ".\\logs.txt")
	viper.SetDefault("SteamID", 0)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err = viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			_, err = os.Create("./config.yaml")
			check(err, false)
			err = viper.WriteConfig()
			check(err, false)
		} else {
			Popup.Alert("Elden Backup", "Unable to read the configuration file. Loading default data. Please, refer to the documentation to solve the issue.")
		}
	}

	return err
}

func OnStartup() {
	c := 0
	exName, err := os.Executable()
	exName = filepath.Base(exName)
	quit = make(chan bool)
	check(err, true)

	// Check if another instance is already running
	pr, err := ps.Processes()
	check(err, true)
	for _, p := range pr {
		if p.Executable() == exName {
			c = c + 1
		}
	}
	if c > 1 {
		Popup.Alert(APP_TITLE, "Another instance of the application is already running.")
		os.Exit(-1)
	}

	// SETUP CONFIG FILES
	ViperSetup()

	// Clear log file
	if viper.GetBool("EnableLogging") {
		if err := os.Truncate(viper.GetString("LogsPath"), 0); err != nil {
			check(err, false)
		}
	}

	SAVE_PATH = viper.GetString("SavefileDirectory")
	if SAVE_PATH == "%appdata%\\EldenRing\\SteamID\\" {
		steamid := GetSteamID()
		if steamid == "0" {
			Popup.Alert(APP_TITLE, "Steam needs to be open to use the SteamID autodetection. Open Steam and run the application again or manually set the SteamID in the configuration files.")
			os.Exit(0)
		}
		SAVE_PATH = strings.Replace(SAVE_PATH, "SteamID", steamid, 1)
		STEAMID = steamid
		if viper.GetBool("EnableLogging") {
			Log(viper.GetString("LogsPath"), "Saves directory found at "+SAVE_PATH+" for steam id "+STEAMID+".", LOG_DEBUG)
		}
	}

	// Check if a save file exists
	var _ os.FileInfo
	_, err = os.Stat(ResolvePath(SAVE_PATH + GetSaveName()))
	if errors.Is(err, os.ErrNotExist) {
		Popup.Alert(APP_TITLE, "No save file was found. Start your adventure and then open "+APP_TITLE+".")
		os.Exit(0)
	}

	if viper.GetBool("EnableLogging") {
		Log(viper.GetString("LogsPath"), "The application has been correctly started.", LOG_DEBUG)
	}

	if viper.GetBool("backuponstartup") {
		BackupFile(SAVE_PATH+GetSaveName(), BCK_STARTUP)
	}

	if viper.GetInt("backupintervaltimeout") > 0 {
		go IntervalledBackup(viper.GetInt("backupintervaltimeout"))
	}
}

func main() {
	OnStartup()
	if viper.GetBool("EnableSaveListener") {
		var err error
		wHandle, err = fsnotify.NewWatcher()
		check(err, true)
		go StartWatcher(wHandle)
	}
	systray.Run(onReady, onExit)
}

func onReady() {
	dat, err := os.ReadFile("./Icon.ico")
	if check(err, false) {
		systray.SetIcon(dat)
	}
	systray.SetTitle(APP_TITLE)
	systray.SetTooltip(APP_TITLE)
	bckMenu := systray.AddMenuItem("Backup Now", "Execute a backup on the spot.")
	systray.AddSeparator()
	quitMenu := systray.AddMenuItem("Quit", "Quit the app.")

	go func() {
		for {
			select {
			case <-bckMenu.ClickedCh:
				if viper.GetBool("EnableLogging") {
					Log(viper.GetString("LogsPath"), "Manual backup requested via system tray.", LOG_DEBUG)
				}
				BackupFile(SAVE_PATH+GetSaveName(), BCK_MANUAL)
			case <-quitMenu.ClickedCh:
				systray.Quit()
			}
		}
	}()

	// Sets the icon of a menu item. Only available on Mac and Windows.
	//mQuit.SetIcon(icon.Data)
}

func onExit() {
	quit <- true
	wHandle.Close()
}
