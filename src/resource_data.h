#ifndef _INDEX_H
#define _INDEX_H

#include <stdint.h>

typedef struct {
  uint8_t offset_lo;
  uint8_t offset_mid;
  uint8_t offset_hi;
  uint8_t size_lo;
  uint8_t size_hi;
} resource_entry_t;

#define RESOURCE_TYPE_LOGIC 0
#define RESOURCE_TYPE_PIC 1
#define RESOURCE_TYPE_VIEW 2
#define RESOURCE_TYPE_SOUND 3

#define RESOURCE_PRESENT(e)                                                   \
  ((e).size_lo != 0xFF || (e).size_hi != 0xFF)

int init_resource_data();

resource_entry_t getResourceIndex(uint8_t type, uint8_t num);

// Reassembles an entry's 3-byte offset into a usable 24-bit value.
unsigned long resource_offset(resource_entry_t e);

// Reassembles an entry's 2-byte size.
unsigned int resource_size(resource_entry_t e);

// seek our data file handle to the specified address
void seek_resource(unsigned long addr);

// load data into xstack and leave it there
int load_xstack(unsigned count);

int data_file();

#endif /* _INDEX_H */
