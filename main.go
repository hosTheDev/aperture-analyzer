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
	"unsafe"
)

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

	fmt.Println("\n[Aperture] Program finished.")
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
		return errors.New("C++ engine reported an error on start")
	}

	fmt.Println("[Aperture] C++ engine acknowledged the request successfully.")
	return nil
}