#include "capture.h" // Include our own header file
#include <pcap.h>    // We still need the Npcap headers for the implementation
#include <iostream>

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