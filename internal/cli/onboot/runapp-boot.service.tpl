[Unit]
Description=Runapp boot
After=network.target

[Service]
ExecStart=$BINARY_PATH onboot
Restart=no
# when onboot process finished do not terminate the started apps
KillMode=none

[Install]
WantedBy=default.target
