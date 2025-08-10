# Aperture Network Analyzer

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![C++](https://img.shields.io/badge/C%2B%2B-00599C?style=for-the-badge&logo=c%2B%2B&logoColor=white)
![Platform](https://img.shields.io/badge/Platform-Windows-0078D6?style=for-the-badge&logo=windows&logoColor=white)

Aperture is a lightweight, high-performance command-line network packet analyzer for Windows, built with a hybrid Go and C++ architecture.

It leverages the raw capture speed of C++ and the Npcap library, combined with the powerful concurrency and parsing capabilities of Go and `gopacket`.

## Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Technology Stack](#technology-stack)
- [Prerequisites & Setup](#prerequisites--setup)
- [Building the Application](#building-the-application)
- [How to Run](#how-to-run)

## Features

-   Lists all available network interfaces on your machine.
-   Captures live packet data from a user-selected interface.
-   Parses Ethernet, IPv4, TCP, and UDP layers.
-   Displays a real-time, formatted summary of captured packets (Timestamp, IP addresses, Ports, Protocol, and Length).
-   Cleanly separates high-performance capture logic (C++) from application control and parsing (Go).

## Architecture

Aperture uses a hybrid model to get the best of both worlds:

-   **Go (Front-end/Controller):** The main application is written in Go. It handles the user interface, manages the application lifecycle, and uses the `gopacket` library to parse and display packet information.
-   **C++ (Back-end/Engine):** A C++ library acts as a high-performance wrapper around the Npcap library. It is responsible for the low-level tasks of finding network devices and capturing raw packet data efficiently. The capture loop runs on a dedicated background thread to keep the Go UI responsive.
-   **cgo (The Bridge):** Go's `cgo` tool is the glue that enables seamless communication between the Go controller and the C++ engine.

## Technology Stack

-   **Go**: Main application logic and user interface.
-   **C++11**: High-performance packet capture engine.
-   **cgo**: Bridge between Go and C++.
-   **Npcap**: The underlying packet capture driver and library for Windows.
-   **MinGW-w64**: The compiler toolchain used for building the C++ code and linking with Go on Windows.
-   **gopacket**: The primary Go library for packet decoding and analysis.

## Prerequisites & Setup

Follow these steps carefully to set up your development environment.

### 1. Install Go

Ensure you have the latest version of Go installed. You can download it from the [official Go website](https://go.dev/dl/).

### 2. Install the C++ Compiler (MinGW-w64)

This project requires the MinGW-w64 GCC toolchain. The easiest way to get it is via **MSYS2**.

1.  **Install MSYS2:** Download and install MSYS2 from [msys2.org](https://www.msys2.org/). Follow their installation instructions.

2.  **Install MinGW-w64:** After installing MSYS2, open the **MSYS2 UCRT64** terminal (from the Start Menu) and run the following command to install the compiler toolchain:
    ```bash
    pacman -S mingw-w64-ucrt-x86_64-gcc
    ```

3.  **Add to PATH:** Add the MinGW-w64 `bin` directory to your Windows System Environment `PATH`. If you used the default MSYS2 installation location, the path will be:
    ```
    C:\msys64\ucrt64\bin
    ```
    *This is a critical step for Go to find the C++ compiler.*

### 3. Download the Npcap SDK

The project needs the Npcap headers and libraries to compile the C++ engine.

1.  **Download:** Get the latest Npcap SDK from the [official Npcap website](https://npcap.com/#download) (e.g., `npcap-sdk-1.13.zip`).
2.  **Place SDK:** Create a folder named `NpCap` in the root of the project directory. Unzip the contents of the SDK into this folder. The final structure should look like this:
    ```
    <project-root>/
    ├── NpCap/
    │   ├── Include/
    │   ├── Lib/
    │   └── ... (other SDK files)
    ├── main.go
    ├── capture.cpp
    ├── capture.h
    └── ...
    ```
    > **Note:** The `NpCap/` directory is listed in `.gitignore` and should not be committed to version control.

## Building the Application

A simple build script is provided. Open your terminal (Command Prompt, PowerShell, or Git Bash) in the project root and run:

```bash
.\build.bat
```

This script will invoke `go build` with the correct flags, compiling the Go and C++ source files and linking them into a single executable named `aperture.exe`.

## How to Run

After a successful build, you can run the application directly from your terminal:

```bash
.\aperture.exe
```

1.  The program will start and display a list of available network devices.
2.  Enter the number corresponding to the device you wish to analyze.
3.  The capture session will begin, and you will see a real-time summary of packets.
4.  To stop the capture, simply **press Enter**. The program will then shut down gracefully.

---
This project is licensed under the MIT License. See the LICENSE file for details.