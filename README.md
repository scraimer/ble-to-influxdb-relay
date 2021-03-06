# Relay from Mi Temperature sensor to InfluxDB on Raspberry Pi Zero

## Get Gatt Dependencies

After cloning this repo, run the following commands:

    cd $CLONEDIR
    go mod edit -replace=github.com/paypal/gatt=./local/gatt
    mkdir local/
    cd local
    git clone https://github.com/paypal/gatt.git
    cd gatt
    git checkout 4ae819d591cfc94c496c45deb55928469542beec
    git apply ../../gatt-changes.patch
    go mod init local/gatt
    go mod tidy
    
When doing `go install` below, there might be some dependencies that need `go get`

## Build

    cd $CLONEDIR
    go install github.com/scraimer/rpi0-ble-relay

## Run

    sudo /home/pi/go/bin/rpi0-ble-relay

## Install to run on boot

    sudo python3 -s install-service.py

