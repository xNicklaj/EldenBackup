package main

//go:generate goversioninfo -icon=./Icon.ico

import (
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
)

func GetSaveName() string {
	if viper.GetBool("UseSeamlessCoop") == true {
		return "ER0000.co2"
	}
	return "ER0000.sl2"
}

func CopyFile(src string, dst string) {
	srcFile, err := os.Open(src)
	check(err)
	defer srcFile.Close()

	destFile, err := os.Create(dst) // creates if file doesn't exist
	check(err)
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile) // check first var for number of bytes copied
	check(err)

	err = destFile.Sync()
	check(err)
}

func check(err error) {
	if err != nil {
		fmt.Printf("Error : %s\n", err.Error())
		os.Exit(1)
	}
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
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()
	if err := w.Add(ResolvePath(SAVE_PATH)); err != nil {
		log.Fatalln(err)
	}
	for path, f := range w.WatchedFiles() {
		fmt.Printf("%s: %s\n", path, f.Name())
	}
	if err := w.Start(time.Millisecond * 100); err != nil {
		log.Fatalln(err)
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
	check(err)

	pr, err := ps.Processes()
	check(err)
	for i, p := range pr {
		if i == i && p.Executable() == exName {
			c = c + 1
		}
	}
	if c > 1 {
		Popup.Alert("Elden Backup", "Another instance of the application is already running.")
		os.Exit(-1)
	}

	viper.SetDefault("BackupDirectory", "%appdata%\\EldenRingBackup\\")
	viper.SetDefault("BackupOnStartup", true)
	viper.SetDefault("BackupIntervalTimeout", 5)
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
	check(err)
	systray.SetIcon(dat)
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
