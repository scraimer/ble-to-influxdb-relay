package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"encoding/binary"
	"math"
	"time"
	"strings"
	"encoding/json"
	"io/ioutil"
	"os"
	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type Config struct {
	Name string	`json:"name"`
}

var config Config

func readConfig(configFilename string) (Config, error) {
	file, err_open := os.Open(configFilename)
	if err_open != nil {
		log.Fatal(err_open)
		return Config{}, err_open
	}
	defer file.Close()
	byteValue, err_read := ioutil.ReadAll(file)
	if err_read != nil {
		log.Fatal(err_read)
		return Config{}, err_read
	}
	var outConfig Config
	err_json := json.Unmarshal(byteValue, &outConfig)
	if err_json != nil {
		log.Fatal(err_json)
		return Config{}, err_json
	}
	return outConfig, nil
}

func onStateChanged(device gatt.Device, s gatt.State) {
	switch s {
	case gatt.StatePoweredOn:
		fmt.Println("Scanning for ATC temperature advertisements...")
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

type measurement struct {
	mac string
	temperature float64
	humidity float64
	battery_percent float64
	battery_mv float64
	frame_packet_counter float64
}

func parseMeasurement(mac string, data []byte) *measurement {
	if data == nil || len(data) != 15 {
		return nil
	}

	mac_lowercase := strings.ToLower(strings.ReplaceAll(mac, ":", ""))

	temperature_int16, err := readInt16(data[8:10])
	if err != nil {
		fmt.Printf("Error parsing temperature: %v\n", err)
		return nil
	}
	temperature := float64(temperature_int16) / 10.0
	humidity := data[10]
	battery_percent := data[11]
	battery_mv := binary.LittleEndian.Uint16(data[12:])
	frame_packet_counter := data[14]

	m := measurement{
		mac: mac_lowercase,
		temperature: float64(temperature),
		humidity: float64(humidity),
		battery_percent: float64(battery_percent),
		battery_mv: float64(battery_mv),
		frame_packet_counter: float64(frame_packet_counter),
	}
	return &m
}

var client influxdb2.Client
var writeAPI api.WriteAPI

func initInflux() {
   //userName := "my-user"
   //password := "my-password"
   //auth_token := fmt.Sprintf("%s:%s",userName, password)
   auth_token := ""
   org_name := ""
   dest_db := "temperature_sensors_v1"
   dest_retention_policy := ""
   db_string := fmt.Sprintf("%s/%s", dest_db, dest_retention_policy)

   // Create a new client using an InfluxDB server base URL and an authentication token
   // and set batch size to 10 
   client = influxdb2.NewClientWithOptions("http://hinge-iot:8086", auth_token,
	   influxdb2.DefaultOptions().SetBatchSize(10))
   // Get non-blocking write client
   writeAPI = client.WriteAPI(org_name, db_string)
}

func closeInflux() {
    // Force all unwritten data to be sent
    writeAPI.Flush()
    // Ensures background processes finishes
    client.Close()
}

func writeMeasurement(m *measurement) {
   // create point
	p := influxdb2.NewPointWithMeasurement("air").
		AddTag("mac", m.mac).
		AddTag("relay", config.Name).
   	AddField("temperature", math.Round(m.temperature*100)/100).
   	AddField("humidity", m.humidity).
   	AddField("battery_percent", m.battery_percent).
   	AddField("battery_mv", m.battery_mv).
   	AddField("frame_packet_counter", m.frame_packet_counter).
		SetTime(time.Now())
   // write asynchronously
   writeAPI.WritePoint(p)
}

func onPeripheralDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	debug := false
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
	m := parseMeasurement(mac, a.ManufacturerData)
	if m != nil {
		if debug {
			fmt.Printf("MAC=%s temperature=%g humidity=%d%%", m.mac, m.temperature, m.humidity)
			fmt.Printf(" battery_percent=%d%% battery_mv=%d", m.battery_percent, m.battery_mv)
			fmt.Printf(" frame_packet_counter=%d\n", m.frame_packet_counter)
		}
		writeMeasurement(m)
	} else {
		if debug {
			fmt.Printf("Invalid 0x181A packet from MAC %s\n", mac)
		}
	}
}

func main() {
	var config_err error
	config, config_err = readConfig("/etc/ble-relay.conf")
	if config_err != nil {
		log.Fatalf("Error reading config: %s\n", config_err)
	}
	initInflux()
	device, err := gatt.NewDevice(option.DefaultClientOptions...)
	if err != nil {
		log.Fatalf("Failed to open device, err: %s\n", err)
		return
	}
	device.Handle(gatt.PeripheralDiscovered(onPeripheralDiscovered))
	device.Init(onStateChanged)
	select {}
}
