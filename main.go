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

	fmt.Println("\n[Aperture] Program finished.")
}