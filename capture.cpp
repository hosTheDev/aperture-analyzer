#include "capture.h" // Include our own header file
#include <pcap.h>    // We still need the Npcap headers for the implementation
#include <iostream>

//todo: Ensure if we really need the following includes.
#include <string>     // For std::string
#include <sstream>    // For std::stringstream
#include <stdlib.h>   // For free()
#include <string.h>   // For strdup()

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

int start_capture_session(const char* device_name) {
    if (!device_name) {
        return -1; // Return error if device_name is null
    }

    // This is our placeholder logic. It just prints the device name it received.
    std::cout << "\n[C++ ENGINE] Instructed to start capture on device: " << device_name << std::endl;
    
    // In the future, this is where the pcap_open_live() logic will go.
    
    return 0; // Return 0 to indicate success
}