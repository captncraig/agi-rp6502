#include "resource_data.h"
#include "error.h"
#include <fcntl.h>
#include <rp6502.h>
#include <stdio.h>
#include <stdint.h>
#include <unistd.h>
#include <time.h>

#define RESOURCE_SLOTS 256

unsigned long resource_offset(resource_entry_t e) {
  return (unsigned long)e.offset_lo | ((unsigned long)e.offset_mid << 8) |
         ((unsigned long)e.offset_hi << 16);
}

unsigned int resource_size(resource_entry_t e) {
  return (unsigned int)e.size_lo | ((unsigned int)e.size_hi << 8);
}

int index_file_handle = -1;
int data_file_handle = -1;
// todo: we may need an additional handle for async stuff like audio
resource_entry_t load_entry;

int init_resource_data(){
  index_file_handle = open("ROM:index.bin", O_RDONLY);
  if (index_file_handle < 0){
    errorf("opening ROM:index.bin: %d\n", index_file_handle);
    return index_file_handle;
  }
  infof("Opened resource index with handle %d\n", index_file_handle);
  data_file_handle = open("ROM:data.bin", O_RDONLY);
  if (data_file_handle < 0){
    errorf("opening ROM:data.bin: %d\n", data_file_handle);
    return data_file_handle;
  }
  infof("Opened resource data with handle %d\n", data_file_handle);
  return 0;
}

static const unsigned int type_index_offset[4] = {
  0 * RESOURCE_SLOTS * sizeof(resource_entry_t),
  1 * RESOURCE_SLOTS * sizeof(resource_entry_t),
  2 * RESOURCE_SLOTS * sizeof(resource_entry_t),
  3 * RESOURCE_SLOTS * sizeof(resource_entry_t),
};

resource_entry_t getResourceIndex(uint8_t type, uint8_t num){
  // calculate offset. each index is 5 bytes.
  // Multiply by 5 with a shift and add.
  unsigned int offset = type_index_offset[type] + (num << 2) + num;

  long seekResult = f_lseek(offset, SEEK_SET, index_file_handle);
  if (seekResult < 0) {
    load_entry.size_hi = 0xFF;
    load_entry.size_lo = 0xFF;
    errorf("seeking in index: %ld\n", seekResult);
  }
  int nRead = read(index_file_handle, &load_entry, sizeof(resource_entry_t));
  if (nRead != sizeof(resource_entry_t)) {
    errorf("reading from index: %d\n", nRead);
  }
  return load_entry;
}

void seek_resource(unsigned long addr){
  if (f_lseek(addr, SEEK_SET, data_file_handle) == -1){
    fatalf("seek error %lx", addr);
  }
}

int load_xstack(unsigned count) {
  ria_push_int(count);
  ria_set_ax(data_file_handle);
  int ax = ria_call_int(RIA_OP_READ_XSTACK);
  return ax;
}

int data_file(){
  return data_file_handle;
}