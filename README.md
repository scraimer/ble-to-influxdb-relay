# Relay from Mi Temperature sensor to InfluxDB on Raspberry Pi Zero

## Get Gatt Dependencies

After cloning this repo, run the following commands:

   cd $CLONEDIR
   mkdir local/
   cd local
   git clone https://github.com/paypal/gatt.git
   cd gatt
   git checkout 4ae819d591cfc94c496c45deb55928469542beec
   git apply ../../gatt-changes.patch

   go mod edit -replace=github.com/paypal/gatt=./local/gatt
   go mod init local/gatt

## Build

   cd $CLONEDIR
   go install github.com/scraimer/rpi0-ble-relay

## Run

   sudo /home/pi/go/bin/rpi0-ble-relay

