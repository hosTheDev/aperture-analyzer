#ifndef CAPTURE_H
#define CAPTURE_H

// #include <pcap.h>    // We still need the Npcap headers for the implementation

// This tells the C++ compiler to make these functions available in a way
// that is compatible with the C language, which is what Go understands.
#ifdef __cplusplus
extern "C" {
#endif

int get_device_count();

#ifdef __cplusplus
}
#endif

#endif // CAPTURE_H