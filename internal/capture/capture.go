package capture

/*
#cgo CXXFLAGS: -I../../NpCap/Include
#cgo LDFLAGS: -L../../NpCap/Lib/x64 -lwpcap -lPacket

#include <stdlib.h>
#include "capture.h"
*/
import "C"

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"unsafe"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// Types are now internal to this package but can be exposed if needed.
type packetData C.PacketData

type DeviceInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GetDevices retrieves the list of network devices from the C++ engine.
func GetDevices() ([]DeviceInfo, error) {
	cJsonString := C.get_all_devices_as_json()
	defer C.free_json_string(cJsonString)
	goJsonString := C.GoString(cJsonString)

	var devices []DeviceInfo
	err := json.Unmarshal([]byte(goJsonString), &devices)
	if err != nil {
		return nil, fmt.Errorf("could not parse JSON from C++ code: %w", err)
	}
	if len(devices) == 0 {
		return nil, errors.New("C++ engine reported 0 network devices")
	}
	return devices, nil
}

// StartCapture tells the C++ engine to start capturing on a given device.
func StartCapture(device DeviceInfo) error {
	cDeviceName := C.CString(device.Name)
	defer C.free(unsafe.Pointer(cDeviceName))

	result := C.start_capture_session(cDeviceName)
	if result != 0 {
		return errors.New("c++ engine reported an error on start")
	}
	return nil
}

// StopCapture tells the C++ engine to stop.
func StopCapture() {
	C.stop_capture()
}

// PollForPackets is a long-running function that retrieves packets and sends their summaries to a channel.
func PollForPackets(packetChan chan<- string) {
	for {
		packet := C.get_next_packet()
		if packet != nil {
			goPacket := (*packetData)(packet)
			goBytes := C.GoBytes(unsafe.Pointer(goPacket.bytes), C.int(goPacket.caplen))
			summary := parsePacket(goBytes)
			packetChan <- summary // Send the summary to the channel
			C.free_packet(packet)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// parsePacket is an internal helper that turns raw bytes into a summary string.
func parsePacket(data []byte) string {
	packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.Default)
	timestamp := time.Now().Format("15:04:05.000000")
	summary := fmt.Sprintf("%s |", timestamp)

	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	udpLayer := packet.Layer(layers.LayerTypeUDP)

	if ipLayer != nil {
		ip, _ := ipLayer.(*layers.IPv4)
		summary += fmt.Sprintf(" %s -> %s |", ip.SrcIP, ip.DstIP)
	} else {
		summary += " Non-IP Packet |"
	}
	if tcpLayer != nil {
		tcp, _ := tcpLayer.(*layers.TCP)
		summary += fmt.Sprintf(" TCP | %s -> %s |", tcp.SrcPort, tcp.DstPort)
	} else if udpLayer != nil {
		udp, _ := udpLayer.(*layers.UDP)
		summary += fmt.Sprintf(" UDP | %s -> %s |", udp.SrcPort, udp.DstPort)
	} else {
		summary += " Non-TCP/UDP |"
	}
	summary += fmt.Sprintf(" Len: %d", len(data))
	return summary
}