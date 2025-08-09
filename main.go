package main

/*
#cgo CXXFLAGS: -I./NpCap/Include
#cgo LDFLAGS: -L./NpCap/Lib/x64 -lwpcap -lPacket

#include <stdlib.h>
#include "capture.h"
*/
import "C"
import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unsafe"

	// Import the gopacket library and its layers package
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// This Go struct now mirrors the C struct in capture.h
type PacketData C.PacketData

type DeviceInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// main is now the high-level coordinator of the application.
func main() {
	fmt.Println("[Aperture] Starting Go program...")

	// Step 1: Get the list of devices from the C++ engine.
	devices, err := getDevices()
	if err != nil {
		fmt.Printf("[Aperture] Error: %v\n", err)
		return
	}

	// Step 2: Display the list and get the user's selection.
	selectedDevice, err := promptForDeviceSelection(devices)
	if err != nil {
		fmt.Printf("[Aperture] Error: %v\n", err)
		return
	}

	// Step 3: Pass the selected device to the C++ engine to start the session.
	err = initiateCapture(selectedDevice)
	if err != nil {
		fmt.Printf("[Aperture] Error: %v\n", err)
		return
	}

	// Goroutine to listen for the "stop" signal from the user
	stopChan := make(chan struct{})
	go func() {
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		close(stopChan)
	}()

	// Loop until the user signals to stop
pollLoop:
	for {
		select {
		case <-stopChan:
			// User pressed Enter
			break pollLoop
		default:
			// Non-blocking check for a new packet
			packet := C.get_next_packet()
			if packet != nil {
				// We got a packet! Process it.
				goPacket := (*PacketData)(packet)
				goBytes := C.GoBytes(unsafe.Pointer(goPacket.bytes), C.int(goPacket.caplen))

				// fmt.Printf("Go received packet! Timestamp: %d.%06d, Size: %d bytes\n",
				// 	goPacket.tv_sec, goPacket.tv_usec, len(goBytes))

				//Parse and print the packet summary
				parseAndPrintPacket(goBytes)

				// CRITICAL: Free the C memory for this packet.
				C.free_packet(packet)
			} else {
				// No packet available, wait briefly to avoid busy-looping
				time.Sleep(10 * time.Millisecond)
			}
		}
	}

	// time.Sleep(2 * time.Second)

	// fmt.Println("\n[Aperture] Capture in progress on background thread...")
	// fmt.Println("[Aperture] Press Enter to stop capture.")
	// bufio.NewReader(os.Stdin).ReadBytes('\n')

	// fmt.Println("[Aperture] Stop signal received. Informing C++ engine...")
	C.stop_capture() // Tell the C++ thread to stop.

	time.Sleep(500 * time.Millisecond)

	fmt.Println("\n[Aperture] Program finished.")
}

// This function uses gopacket to parse and print packet details.
func parseAndPrintPacket(data []byte) {
	// Decode a packet
	packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.Default)

	// Get the major protocol layers
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	udpLayer := packet.Layer(layers.LayerTypeUDP)
	
	timestamp := time.Now().Format("15:04:05.000000")
	summary := fmt.Sprintf("%s |", timestamp)

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
	fmt.Println(summary)
}

// getDevices handles the C++ call to retrieve the device list as JSON.
func getDevices() ([]DeviceInfo, error) {
	fmt.Println("[Aperture] Calling C++ engine to get network devices...")
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

// promptForDeviceSelection displays the devices and handles the user input loop.
func promptForDeviceSelection(devices []DeviceInfo) (DeviceInfo, error) {
	fmt.Printf("\n[Aperture] Success! Found %d network devices:\n", len(devices))
	fmt.Println("-------------------------------------------------")
	for i, device := range devices {
		fmt.Printf("  %d: %s\n     %s\n", i+1, device.Name, device.Description)
	}
	fmt.Println("-------------------------------------------------")

	for {
		fmt.Printf("\nEnter the number of the device to start analysis (1-%d): ", len(devices))
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		choice, err := strconv.Atoi(input)
		if err == nil && choice >= 1 && choice <= len(devices) {
			selectedIndex := choice - 1
			return devices[selectedIndex], nil // Return the selected device
		}
		fmt.Printf("Invalid input. Please enter a number between 1 and %d.\n", len(devices))
	}
}

// initiateCapture handles the Go-to-C call to start the capture session.
func initiateCapture(device DeviceInfo) error {
	fmt.Printf("\n[Aperture] You selected: %s\n", device.Name)
	fmt.Println("[Aperture] Passing selection to C++ engine...")

	cDeviceName := C.CString(device.Name)
	defer C.free(unsafe.Pointer(cDeviceName))

	result := C.start_capture_session(cDeviceName)
	if result != 0 {
		return errors.New("c++ engine reported an error on start")
	}

	fmt.Println("[Aperture] C++ engine acknowledged the request successfully.")
	return nil
}
