# runapp - Run and manage background processes (apps)

* 🧑‍🔧 DX focused
* 🌱 Lightweight footprint
* 📄 Integrated log viewer
* 🔌 Starts on boot _(only for 🐧 Linux via systemd)_
* 🐧 Linux support (MacOS is experimental)

[![asciicast](https://asciinema.org/a/se9dTCtVJJ0hyXkclU7kFSY5C.svg)](https://asciinema.org/a/se9dTCtVJJ0hyXkclU7kFSY5C?speed=2)

## Install
```shell
curl -sSL https://raw.githubusercontent.com/0xB1a60/runapp/main/install.sh | bash
```

## Usage
All commands support easy to use Terminal User Interface 🧙

* `runapp` - List all apps
* `runapp run` - Run an app
* `runapp restart` - Restart an app
* `runapp status` - Read the status of an app
* `runapp logs` - Stream the logs (stdout,stderr) of an app
* `runapp kill` - Kill an app
* `runapp remove` - Remove an app
* `runapp install-onboot` - Set up a systemd service to automatically start `runapp` at boot

## Other
Inspired by [hapless](https://github.com/bmwant/hapless)

[MIT License](https://raw.githubusercontent.com/0xB1a60/runapp/main/LICENSE)
