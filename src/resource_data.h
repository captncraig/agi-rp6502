#ifndef _INDEX_H
#define _INDEX_H

typedef struct {
  unsigned char offset_lo;
  unsigned char offset_mid;
  unsigned char offset_hi;
  unsigned char size_lo;
  unsigned char size_hi;
} resource_entry_t;

#define RESOURCE_TYPE_LOGIC 0
#define RESOURCE_TYPE_PIC 1
#define RESOURCE_TYPE_VIEW 2
#define RESOURCE_TYPE_SOUND 3

#define RESOURCE_ADDR ((unsigned char *)0x8000)

#define RESOURCE_PRESENT(e)                                                   \
  ((e).size_lo != 0xFF || (e).size_hi != 0xFF)

int init_resource_data();

resource_entry_t getResourceIndex(unsigned char type, unsigned char num);

// Reassembles an entry's 3-byte offset into a usable 24-bit value.
unsigned long resource_offset(resource_entry_t e);

// Reassembles an entry's 2-byte size.
unsigned int resource_size(resource_entry_t e);

// load a resource into banked memory, returning the bank number it was loaded into
unsigned char load_resource(unsigned char type, unsigned char num);

#endif /* _INDEX_H */
