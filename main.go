package main

import (
	"fmt"
	"log"
	"time"

    // CORRECTED IMPORT PATHS
	"github.com/hosTheDev/aperture-analyzer/internal/capture"
	"github.com/hosTheDev/aperture-analyzer/internal/tui"
)

func main() {
	// 1. Initialize the low-level capture layer to get devices.
	devices, err := capture.GetDevices()
	if err != nil {
		log.Fatalf("[Aperture] Error: %v\n", err)
	}

	// 2. Use the TUI's prompt to ask the user for a selection.
	selectedDevice, err := tui.PromptForDeviceSelection(devices)
	if err != nil {
		log.Fatalf("[Aperture] Error: %v\n", err)
	}

	// 3. Tell the capture layer to start its work.
	err = capture.StartCapture(selectedDevice)
	if err != nil {
		log.Fatalf("[Aperture] Error: %v\n", err)
	}

	// 4. Create a channel to pass packet data from the capture layer to the TUI.
	packetChan := make(chan capture.PacketInfo)

	// 5. Start polling for packets in a separate goroutine.
	go capture.PollForPackets(packetChan)

	// 6. Start the TUI. This is a blocking call.
	tui.Start(packetChan)

	// 7. Once the TUI is closed, stop the capture and clean up.
	fmt.Println("\n[Aperture] Stop signal received. Informing C++ engine...")
	capture.StopCapture()
	time.Sleep(500 * time.Millisecond)
	fmt.Println("[Aperture] Program finished.")
}