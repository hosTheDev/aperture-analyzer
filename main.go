package main

/*
// These are your existing cgo directives. They tell Go how to find
// the Npcap SDK located inside your project folder. We keep them exactly as-is.
#cgo CXXFLAGS: -I./NpCap/Include
#cgo LDFLAGS: -L./NpCap/Lib/x64 -lwpcap -lPacket

// We include the C header files needed for our C function calls.
#include <stdlib.h>
#include "capture.h"
*/
import "C"
import (
	"encoding/json"
	"fmt"

	"bufio"
	"os"
	"strconv"
	"strings"
	"unsafe"
)

// DeviceInfo is a Go struct that will hold the data for a single device.
// The `json:"..."` tags tell the json package how to map
// the JSON keys from our C++ string to the fields of this Go struct.
type DeviceInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func main() {
	fmt.Println("[Aperture] Starting Go program...")
	fmt.Println("[Aperture] Calling C++ engine to get network devices...")

	// We are choosing to call the new JSON function to get the full device list.
	// The old get_device_count() function is still available, but we are not calling it here.
	cJsonString := C.get_all_devices_as_json()

	// CRITICAL: Defer the call to free the C memory.
	// This guarantees that the memory allocated in C++ gets freed
	// when this function exits, even if there's a panic.
	defer C.free_json_string(cJsonString)

	// Convert the C string to a Go string.
	goJsonString := C.GoString(cJsonString)

	// Unmarshal the JSON string into a slice of our Go structs.
	var devices []DeviceInfo
	err := json.Unmarshal([]byte(goJsonString), &devices)
	if err != nil {
		fmt.Println("[Aperture] Error: Could not parse JSON from C++ code.", err)
		return
	}

	// Check the results and print them.
	if len(devices) == 0 {
		fmt.Println("[Aperture] C++ engine reported 0 network devices.")
	} else {
		fmt.Printf("\n[Aperture] Success! Found %d network devices:\n", len(devices))
		fmt.Println("-------------------------------------------------")
		for i, device := range devices {
			fmt.Printf("  %d: %s\n     %s\n", i+1, device.Name, device.Description)
		}
		fmt.Println("-------------------------------------------------")
	}

	var selectedDevice DeviceInfo
	for {
		fmt.Printf("\nEnter the number of the device to start analysis (1-%d): ", len(devices))
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		choice, err := strconv.Atoi(input)
		if err == nil && choice >= 1 && choice <= len(devices) {
			selectedIndex := choice - 1
			selectedDevice = devices[selectedIndex]
			break // Exit loop on valid selection
		}
		fmt.Printf("Invalid input. Please enter a number between 1 and %d.\n", len(devices))
	}
	
	fmt.Printf("\n[Aperture] You selected: %s\n", selectedDevice.Name)
	fmt.Println("[Aperture] Passing selection to C++ engine...")
	
	// Part 3: New - Bridge from Go to C++
	cDeviceName := C.CString(selectedDevice.Name)
	defer C.free(unsafe.Pointer(cDeviceName)) // Must free memory allocated by C.CString

	result := C.start_capture_session(cDeviceName)
	if result == 0 {
		fmt.Println("[Aperture] C++ engine acknowledged the request successfully.")
	} else {
		fmt.Println("[Aperture] C++ engine reported an error.")
	}

	fmt.Println("\n[Aperture] Program finished.")
}