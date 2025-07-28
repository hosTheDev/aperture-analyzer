// We need this define for Npcap to work correctly on Windows
#define HAVE_REMOTE

// Include the main pcap header file
#include <pcap.h>
// Include the standard I/O header for printing to the console
#include <stdio.h>

int main() {
    // cout << "main running." << endl;
    printf("Npcap Example: Listing Network Devices\n");
    pcap_if_t *alldevs;
    pcap_if_t *d;
    int i = 0;
    char errbuf[PCAP_ERRBUF_SIZE];

    // Retrieve the list of network devices on your machine
    if (pcap_findalldevs_ex(PCAP_SRC_IF_STRING, NULL, &alldevs, errbuf) == -1) {
        // If there's an error, print it
        fprintf(stderr, "Error in pcap_findalldevs_ex: %s\n", errbuf);
        return 1;
    }

    // Loop through the list and print each device's name and description
    for (d = alldevs; d != NULL; d = d->next) {
        printf("%d. %s", ++i, d->name);
        if (d->description) {
            printf(" (%s)\n", d->description);
        } else {
            printf(" (No description available)\n");
        }
    }

    // If no devices were found, print a message
    if (i == 0) {
        printf("\nNo interfaces found! Make sure Npcap is installed correctly.\n");
        return 1;
    }

    // Clean up and free the memory used by the device list
    pcap_freealldevs(alldevs);

    return 0;
}