#include "capture.h" // Include our own header file
#include <pcap.h>    // We still need the Npcap headers for the implementation
#include <iostream>

//todo: Ensure if we really need the following includes.
#include <string>     // For std::string
#include <sstream>    // For std::stringstream
#include <stdlib.h>   // For free()
#include <string.h>   // For strdup()
#include <thread>     // For std::thread
#include <atomic>     // For std::atomic
#include <mutex>      // For std::mutex
#include <queue>      // For std::queue

#pragma region VALUES
static pcap_t* capture_handle = nullptr; // Placeholder for the capture handle
static std::thread capture_thread; // Thread for capturing packets
static std::atomic<bool> is_capturing(false); // Atomic flag to control capturing state
static std::queue<PacketData*> packet_queue; // Queue to hold captured packets
static std::mutex queue_mutex; // Mutex to protect access to the packet queue
#pragma endregion

// This is the actual implementation of the function declared in capture.h
int get_device_count() {
    pcap_if_t *alldevs;
    char errbuf[PCAP_ERRBUF_SIZE];
    int count = 0;

    if (pcap_findalldevs_ex(PCAP_SRC_IF_STRING, NULL, &alldevs, errbuf) == -1) {
        std::cerr << "Error in pcap_findalldevs_ex: " << errbuf << std::endl;
        return -1; // Return -1 on error
    }

    for (pcap_if_t *d = alldevs; d != NULL; d = d->next) count++;

    pcap_freealldevs(alldevs);
    return count;
}

// Helper function to escape special JSON characters.
// This is crucial to prevent creating invalid JSON.
std::string escape_json(const char* s) {
    std::stringstream escaped;
    while (*s) {
        switch (*s) {
            case '"':  escaped << "\\\""; break;
            case '\\': escaped << "\\\\"; break;
            case '\b': escaped << "\\b"; break;
            case '\f': escaped << "\\f"; break;
            case '\n': escaped << "\\n"; break;
            case '\r': escaped << "\\r"; break;
            case '\t': escaped << "\\t"; break;
            default:   escaped << *s; break;
        }
        s++;
    }
    return escaped.str();
}

char* get_all_devices_as_json() {
    pcap_if_t *alldevs;
    char errbuf[PCAP_ERRBUF_SIZE];
    std::stringstream json_stream;
    
    if (pcap_findalldevs_ex(PCAP_SRC_IF_STRING, NULL, &alldevs, errbuf) == -1) {
        // On error, return an empty JSON array
        const char* empty_array = "[]";
        // strdup allocates memory using malloc and copies the string.
        // It's a C-style function perfect for cgo.
        return strdup(empty_array);
    }

    json_stream << "[";
    bool first = true;
    for (pcap_if_t *d = alldevs; d != NULL; d = d->next) {
        if (!first) {
            json_stream << ",";
        }
        first = false;
        json_stream << "{";
        json_stream << "\"name\":\"" << escape_json(d->name) << "\",";
        // Ensure description is not null before using it
        json_stream << "\"description\":\"" << (d->description ? escape_json(d->description) : "") << "\"";
        json_stream << "}";
    }
    json_stream << "]";

    // We are done with the Npcap device list, so free it immediately.
    pcap_freealldevs(alldevs);
    std::string final_json = json_stream.str();
    // Return a C-style string on the heap that Go can later free.
    return strdup(final_json.c_str());
}

void free_json_string(char* json_string) {
    // This function can be called from Go to free the memory
    // allocated by strdup() in the function above.
    free(json_string);
}

// This is the callback function that pcap_loop will call for each packet.
void packet_handler(u_char* user, const struct pcap_pkthdr* header, const u_char* bytes) {
    if(!is_capturing) return; // If capturing is stopped, do nothing.
    
    PacketData *packet = new PacketData();
    packet->len = header->len; // Captured length
    packet->caplen = header->caplen; // Original length 
    packet->tv_sec = header->ts.tv_sec; // Timestamp seconds
    packet->tv_usec = header->ts.tv_usec; // Timestamp microseconds
    packet->bytes = new unsigned char[packet->caplen]; // Allocate memory for the packet data
    memcpy(packet->bytes, bytes, packet->caplen); // Copy the raw packet data

    // Lock the queue mutex to safely add the packet to the queue
    std::lock_guard<std::mutex> lock(queue_mutex);
    packet_queue.push(packet); // Add the packet to the queue
}

// get_next_packet is called by Go to consume data from the queue.
PacketData* get_next_packet() {
    std::lock_guard<std::mutex> lock(queue_mutex);
    if (packet_queue.empty()) {
        return nullptr; // No packets available
    }
    PacketData* packet = packet_queue.front();
    packet_queue.pop();
    return packet;
}

// free_packet allows Go to free the memory we allocated in C++.
void free_packet(PacketData* packet) {
    if (packet) {
        delete[] packet->bytes; // Free the byte array
        delete packet;         // Free the struct itself
    }
}

void capture_worker(std::string device_name){
    char errbuf[PCAP_ERRBUF_SIZE];
    capture_handle = pcap_open_live(device_name.c_str(), BUFSIZ, 1, 1000, errbuf);

    if (capture_handle == nullptr) {
        std::cerr << "[C++ ERROR] Could not open device " << device_name << ": " << errbuf << std::endl;
        is_capturing = false;
        return;
    }

    // Start the capture loop. This is a blocking call.
    // It will only exit when pcap_breakloop() is called or an error occurs.
    pcap_loop(capture_handle, -1, packet_handler, NULL);

    // Cleanup after the loop is broken.
    pcap_close(capture_handle);
    capture_handle = nullptr;
    std::cout << "[C++ THREAD] Capture loop finished.\n";
}

int start_capture_session(const char* device_name) {
    if (!device_name || is_capturing) {
        return -1; // Return error if device_name is null
    }

    // This is our placeholder logic. It just prints the device name it received.
    std::cout << "\n[C++ ENGINE] Instructed to start capture on device: " << device_name << std::endl;
    
    is_capturing = true; // Set the capturing flag to true`

    capture_thread = std::thread(capture_worker, std::string(device_name));
    capture_thread.detach();

    std::cout << "[C++ ENGINE] Capture thread launched for device: " << device_name << std::endl;
    return 0; // Return 0 to indicate success
}

void stop_capture() {
    if (!is_capturing) {
        return;
    }
    
    is_capturing = false;
    if (capture_handle != nullptr) {
        // This tells pcap_loop to stop.
        pcap_breakloop(capture_handle);
    }
}