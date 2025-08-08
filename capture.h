#ifndef CAPTURE_H
#define CAPTURE_H

// #include <pcap.h>    // We still need the Npcap headers for the implementation

// This tells the C++ compiler to make these functions available in a way
// that is compatible with the C language, which is what Go understands.
#ifdef __cplusplus
extern "C" {
#endif

int get_device_count();

// Returns a JSON string representing a list of network devices.
// The caller is responsible for freeing this string using free_json_string().
char* get_all_devices_as_json();

// Frees the memory allocated by get_all_devices_as_json().
void free_json_string(char* json_string);

// New function to start a capture session.
// Returns 0 on success, -1 on failure.
int start_capture_session(const char* device_name);

#ifdef __cplusplus
}
#endif

#endif // CAPTURE_H