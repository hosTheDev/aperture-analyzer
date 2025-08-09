#ifndef CAPTURE_H
#define CAPTURE_H

// This struct will hold the packet data passed from C++ to Go.
// We use fixed-size types to ensure compatibility.
typedef struct PacketData {
    unsigned int len;           // Captured length
    unsigned int caplen;        // Original length
    long long tv_sec;           // Timestamp seconds
    long long tv_usec;          // Timestamp microseconds
    unsigned char* bytes;       // The raw packet data
} PacketData ;


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

// Starts the capture session on a background thread.
int start_capture_session(const char* device_name);

// Signals the background thread to stop capturing.
void stop_capture();

PacketData* get_next_packet();
void free_packet(PacketData* packet);

#ifdef __cplusplus
}
#endif

#endif // CAPTURE_H