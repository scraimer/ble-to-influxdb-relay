package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"encoding/binary"

"github.com/paypal/gatt"
"github.com/paypal/gatt/examples/option"
)

func onStateChanged(device gatt.Device, s gatt.State) {
	switch s {
	case gatt.StatePoweredOn:
		fmt.Println("Scanning for iBeacon Broadcasts...")
		device.Scan([]gatt.UUID{}, true)
		return
	default:
		device.StopScanning()
	}
}

func printHex(a []byte) string {
	out := "["
	for i := range a {
		out += fmt.Sprintf("%02x ", a[i])
	}
	out += "]"
	return out
}

func isTemperatureData(a []byte) bool {
	if a == nil || len(a) < 2 {
		return false
	}
	msg_type := binary.LittleEndian.Uint16(a[0:])
	return (msg_type == 0x181A)
}

func readInt16(src []byte) (int16, error) {
	if src == nil || len(src) < 2 {
		return 0, errors.New("Source of int16 must be []byte of length 2 or more")
	}
	var v int16
	buf := bytes.NewReader(src)
	err := binary.Read(buf, binary.BigEndian, &v)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
		return 0, err
	}
	return v, nil
}

func onPeripheralDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	debug := true
	if a == nil || !isTemperatureData(a.ManufacturerData) {
		return
	}
	//fmt.Println("onPeripheralDiscovered RSSI=" , rssi , " a = " , a, " p = ", p)
	mac := p.ID()
	if debug {
		fmt.Printf("MAC=%s data=%s|%s|%s len=%d\n",
			mac,
			printHex(a.ManufacturerData[0:2]),
			printHex(a.ManufacturerData[2:8]),
			printHex(a.ManufacturerData[8:]),
			len(a.ManufacturerData))
	}
	if len(a.ManufacturerData) == 15 {
		temperature_int16, err := readInt16(a.ManufacturerData[8:10])
		if err != nil {
			fmt.Printf("Error parsing temperature: %v\n", err)
			return
		}
		fmt.Printf("temperature_raw = %s %x\n", printHex(a.ManufacturerData[8:10]), temperature_int16)
		temperature := float64(temperature_int16) / 10.0
		humidity := a.ManufacturerData[10]
		battery_percent := a.ManufacturerData[11]
		battery_mv := binary.LittleEndian.Uint16(a.ManufacturerData[12:])
		frame_packet_counter := a.ManufacturerData[14]
		fmt.Printf("MAC=%s temperature=%g humidity=%d%%", mac, temperature, humidity)
		fmt.Printf(" battery_percent=%d%% battery_mv=%d", battery_percent, battery_mv)
		fmt.Printf(" frame_packet_counter=%d\n", frame_packet_counter)
	} else {
		fmt.Printf("Invalid 0x181A packet from MAC %s\n", mac)
	}

	/*
	b, err := NewiBeacon(a.ManufacturerData)
	if err == nil {
		fmt.Println("UUID: ", b.uuid)
		fmt.Println("Major: ", b.major)
		fmt.Println("Minor: ", b.minor)
		fmt.Println("RSSI: ", rssi)
	}
	*/
}

func main() {
	device, err := gatt.NewDevice(option.DefaultClientOptions...)
	if err != nil {
		log.Fatalf("Failed to open device, err: %s\n", err)
		return
	}
	device.Handle(gatt.PeripheralDiscovered(onPeripheralDiscovered))
	device.Init(onStateChanged)
	select {}
}
