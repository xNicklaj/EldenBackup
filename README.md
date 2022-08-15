<a name="readme-top"></a>
<div align="center">
  <a href="https://github.com/xNicklaj/EldenBackup">
    <img src="Icon.ico" alt="Logo" width="80" height="80">
  </a>

  <h3 align="center">Elden Backup</h3>
  <span align="center">
  </span>
  
  <p align="center">
    Elden Backup is a tool that allows you to backup your Elden Ring saves automatically, without having to worry about losing your progress.<br/><br/>
    <img src="https://github.com/xNicklaj/EldenBackup/actions/workflows/go.yml/badge.svg" alt="Passing action" /> <img src="https://img.shields.io/github/go-mod/go-version/xNicklaj/EldenBackup/main" alt="Go version" /> <img src="https://img.shields.io/github/last-commit/xNicklaj/EldenBackup" alt="Last Commit"/> <img src="https://img.shields.io/github/downloads/xNicklaj/EldenBackup/total" alt="Downloads" /><br/>
    <a href="https://github.com/xNicklaj/EldenBackup/issues">Report Bug</a>
    ·
    <a href="https://github.com/xNicklaj/EldenBackup/issues">Request Feature</a>
  </p>
</div>

## Getting Started
### Prerequisites

None. I've developed this app in Go so that there were no prerequisites at all, except the game itself of course.

### Installation

1. Head to the [actions](https://github.com/xNicklaj/EldenBackup/actions) tab and click on the most recent build.
2. Scroll down to the artifacts section and download EldenBackup.zip.
3. Inside the zip you will find EldenBackup.exe and Icon.ico. Extract them both in your Elden Ring folder (or really any folder at all).
4. Execute EldenBackup.exe and enjoy your game.

Update: you can now download the latest release in the [release tab](https://github.com/xNicklaj/EldenBackup/releases/latest).

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Usage

When you will open the executable, an icon will popup in the system tray, at the bottom-right of you screen. Right clicking on it will allow you to either close the app or execute a backup immediately.

There are three different types of backup:
 - **A**utomatic backup, this happens while you play, and at the end of your session you will have one of these backups per day. If your game crashes and somethings gets corrupted, this is the type of backup you can refer to.
 - **M**anual backup, this backup happens only when you click on "Backup now" from the system tray.
 - **S**tartup backup, this backup happens every time you open EldenBackup.
 - **T**imeout backup, this backup happens every once in a while when you play. By default, it happens every five minutes, but you will be able to change the amount of time between these backups from the configuration file. 

 In your backup folder you'll find all your backups in the format: `ER0000-steamid-timestampK` where K indicates the type of the backup.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

 ## Configuration

 Upon your first launch of Elden Backup, you will find that a configuration file has been created in its containing folder.
 What follows is a list of all the different configuration options:
 
- `BackupDirectory`, this is the directory in which all your saves will be backed up. By default this will be `'%appdata%/EldenRingBackup'`.
- `BackupIntervalTimeout`, this is the time in minutes between the timeout backups. By default five minutes, `10`.
- `BackupOnStartup`, this specifies whether Elden Backup will perform a backup upon startup. By default `true`.
- `LimitTimeoutBackups`, this option allows Elden Backup to only keep a certain number of timeout backups in the backup folder when set to a value greater than zero. By default this is set to `0`.
- `LimitAutoBackups`, this option allows Elden Backup to only keep a certain number of auto backups in the backup folder when set to a value greater than zero. By default this is set to `0`.
- `UseSeamlessCoop`, set this to true if you're using the [SeamlessCoop](https://www.nexusmods.com/eldenring/mods/510) mod by [LukeYui](https://www.nexusmods.com/eldenring/users/49594931?tab=about+me). This app was made with SeamlessCoop in mind, as some of its mechanics might break quest triggers, therefore this is set to `true` by default.
- `SavefileDirectory`, set this to your EldenRing save path. The application should be able to find the savefiles automatically, but in case it doesn,'t, this is what you can change. Defaults at `'%appdata%\EldenRing\SteamId\'`.
- `SteamId`, set this to your SteamId if you have multiple steam accounts logged in your installation. This is enabled only if SavefileDirectory is kept at its default value. The default value of SteamId is `0`.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Restore a backup

To restore a backup, all you have to do is to copy-paste it inside `%appdata%/EldenRing` and rename it `ER0000.co2` or `ER0000.sl2` depending on whether you're using the SeamlessCoop mod or not.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Roadmap

- [x] Basic backup functionalities.
- [x] Timeout backups.
- [x] Save space by deleting older backups.
- [x] Automatic detection of the SteamID.
- [ ] Optionally launch Elden Ring or the Seamless Coop launcher automatically.
- [ ] Add EXE metadata.

You can check the full changelog <a href="https://github.com/xNicklaj/EldenBackup/releases/">here</a>.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- LICENSE -->
## License

Distributed under the GNU GPLv3 License. See `LICENSE.txt` for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Acknowledgements

 - App Icon by [Brastertag](https://www.deviantart.com/brastertag/art/Elden-Ring-919397405), on Deviantart.
 - Developed in [Golang](https://go.dev).
 - This app is in no way affiliated or endorsed by FromSoftware.

<p align="right">(<a href="#readme-top">back to top</a>)</p>
