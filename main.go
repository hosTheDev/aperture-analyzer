package main

/*
// CFLAGS is for C files.
#cgo CFLAGS: -I./NpCap/Include

// CXXFLAGS is for C++ files. We need this for capture.cpp
#cgo CXXFLAGS: -I./NpCap/Include

// LDFLAGS tells the linker what libraries to use.
#cgo LDFLAGS: -L./NpCap/Lib/x64 -lwpcap -lPacket

// We just include our header file.
#include "capture.h"
*/
import "C" // This enables cgo
import "fmt"

func main() {
    fmt.Println("Go program started. Calling C code...")

    // Call the C function, which is available because we included the header.
    deviceCount := C.get_device_count()

    if deviceCount < 0 {
        fmt.Println("Error: Could not get device list from C code.")
    } else {
        fmt.Printf("Success! C code reported finding %d network devices.\n", deviceCount)
    }

    fmt.Println("Go program finished.")
}