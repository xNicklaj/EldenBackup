package main

//go:generate goversioninfo -icon=./Icon.ico

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ChromeTemp/Popup"
	"github.com/getlantern/systray"
	"github.com/mitchellh/go-ps"
	"github.com/radovskyb/watcher"
	"github.com/spf13/viper"
)

var wHandle *watcher.Watcher
var quit chan bool

const APP_TITLE = "Elden Backup"
const SAVE_PATH = "%appdata%\\EldenRing\\"

const (
	BCK_STARTUP int = 0
	BCK_MANUAL  int = 1
	BCK_AUTO    int = 2
	BCK_TIMEOUT int = 3
)

func GetSaveName() string {
	if viper.GetBool("UseSeamlessCoop") {
		return "ER0000.co2"
	}
	return "ER0000.sl2"
}

func CopyFile(src string, dst string) {
	srcFile, err := os.Open(src)
	check(err, true)
	defer srcFile.Close()

	destFile, err := os.Create(dst) // creates if file doesn't exist
	check(err, true)
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile) // check first var for number of bytes copied
	check(err, true)

	err = destFile.Sync()
	check(err, true)
}

func check(err error, exit bool) bool {
	if err != nil {
		fmt.Printf("Error : %s\n", err.Error())
		if exit {
			os.Exit(1)
		}
		return false
	}
	return true
}

func ResolvePath(path string) string {
	return strings.Replace(path, "%appdata%", os.Getenv("APPDATA"), -1)
}

func BackupFile(inp_path string, mode int) {
	bck_path := ResolvePath(viper.GetString("BackupDirectory"))
	file_path := ResolvePath(inp_path)

	ext := filepath.Ext(file_path)
	filename := filepath.Base(file_path)
	filename = strings.Trim(filename, ext)

	ctime := time.Now()

	switch mode {
	case BCK_STARTUP:
		bck_path = bck_path + filename + "-" + ctime.Format("20060102_1504") + "S" + ext
	case BCK_AUTO:
		bck_path = bck_path + filename + "-" + ctime.Format("20060102") + "A" + ext
	case BCK_MANUAL:
		bck_path = bck_path + filename + "-" + ctime.Format("20060102_1504") + "M" + ext
	case BCK_TIMEOUT:
		bck_path = bck_path + filename + "-" + ctime.Format("20060102_1504") + "T" + ext
	}
	CopyFile(file_path, bck_path)
}

func StartWatcher(w *watcher.Watcher) {
	w.FilterOps(watcher.Write)
	go func() {
		for {
			select {
			case event := <-w.Event:
				fmt.Println(event) // Print the event's info.
				if filepath.Base(event.Path) == GetSaveName() {
					BackupFile(event.Path, BCK_AUTO)
				}
			case err := <-w.Error:
				log.Fatalln(err, true)
			case <-w.Closed:
				return
			}
		}
	}()
	if err := w.Add(ResolvePath(SAVE_PATH)); err != nil {
		log.Fatalln(err, true)
	}
	for path, f := range w.WatchedFiles() {
		fmt.Printf("%s: %s\n", path, f.Name())
	}
	if err := w.Start(time.Millisecond * 100); err != nil {
		log.Fatalln(err, true)
	}
}

func IntervalledBackup(delay int) {
	ticker := time.NewTicker(time.Duration(delay) * time.Minute)
	for {
		select {
		case <-ticker.C:
			BackupFile(SAVE_PATH+GetSaveName(), BCK_MANUAL)
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func OnStartup() {
	c := 0
	exName, err := os.Executable()
	exName = filepath.Base(exName)
	quit = make(chan bool)
	check(err, true)

	// Check if a save file exists
	var _ os.FileInfo
	_, err = os.Stat(ResolvePath(SAVE_PATH + GetSaveName()))
	if errors.Is(err, os.ErrNotExist) {
		Popup.Alert(APP_TITLE, "No save file was found. Start your adventure and then open "+APP_TITLE+".")
		os.Exit(0)
	}

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
	viper.SetDefault("BackupDirectory", "%appdata%\\EldenRingBackup\\")
	viper.SetDefault("BackupOnStartup", true)
	viper.SetDefault("BackupIntervalTimeout", 10)
	viper.SetDefault("UseSeamlessCoop", true)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			os.Create("./config.yaml")
			viper.WriteConfig()
		} else {
			Popup.Alert("Elden Backup", "Unable to read the configuration file. Loading default data. Please, refer to the documentation to solve the issue.")
		}
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
	wHandle = watcher.New()
	go StartWatcher(wHandle)
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
