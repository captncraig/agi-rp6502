#ifndef _INDEX_H
#define _INDEX_H

#include <stdint.h>

typedef struct {
  uint32_t offset;
  uint16_t size;
  uint8_t current_bank;
  uint8_t padding;
} resource_entry_t;

typedef struct {
  resource_entry_t logics[256];
  resource_entry_t pics[256];
  resource_entry_t views[256];
  resource_entry_t sounds[256];
} resource_index_t;

#define resource_index (*(resource_index_t *)0x8000)

#define RESOURCE_TYPE_LOGIC 0
#define RESOURCE_TYPE_PIC 1
#define RESOURCE_TYPE_VIEW 2
#define RESOURCE_TYPE_SOUND 3

#define RESOURCE_PRESENT(e)                                                   \
  ((e).size != 0xFFFF)

int init_resource_data();

// seek our data file handle to the specified address
void seek_resource(unsigned long addr);
int data_file();

#endif /* _INDEX_H */
