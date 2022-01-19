#!/usr/bin/python3

from pathlib import Path
from textwrap import dedent
import subprocess


def create_service_script(name:str) -> Path:
    DEST = Path("/usr/bin")
    content = dedent("""\
        DATE=`date '+%Y-%m-%d %H:%M:%S'`
        echo "BLE Relay service started at ${DATE}" | systemd-cat -p info

        sudo /home/pi/go/bin/rpi0-ble-relay
    """)
    out = DEST / f"{name}_service.sh"
    out.write_text(content, encoding="utf-8")
    out.chmod(755)
    return out


def create_systemd_unit_file(script:Path, name:str) -> Path:
    DEST = Path("/lib/systemd/system")
    content = dedent(f"""\
        [Unit]
        Description=Bluetooth Low-Energy relay of sensor data to InfluxDB
        After=network-online.target bluetooth.service
        Wants=network-online.target bluetooth.service systemd-networkd-wait-online.service
        StartLimitIntervalSec=500
        StartLimitBurst=5

        [Service]
        Type=simple
        ExecStart=/bin/bash {str(script)}
        Restart=on-failure
        RestartSec=60s

        [Install]
        WantedBy=multi-user.target
    """)
    out = DEST / f"{name}.service"
    out.write_text(content, encoding="utf-8")
    out.chmod(644)
    return out


def enable_and_start_service(service_name:str) -> None:
    subprocess.run(f"systemctl enable {service_name}", shell=True)
    subprocess.run(f"systemctl start {service_name}", shell=True)


def main() -> None:
    name = "ble-relay"
    script = create_service_script(name=name)
    create_systemd_unit_file(script=script, name=name)
    enable_and_start_service(service_name=name)


if __name__ == "__main__":
    main()
