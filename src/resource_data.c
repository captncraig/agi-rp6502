#include "resource_data.h"
#include "error.h"
#include "banking.h"
#include <fcntl.h>
#include <rp6502.h>
#include <stdio.h>
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

resource_entry_t getResourceIndex(unsigned char type, unsigned char num){
  long offset = (long)type * RESOURCE_SLOTS * sizeof(resource_entry_t) +
                (long)num * sizeof(resource_entry_t);
  long errno = f_lseek(offset, SEEK_SET, index_file_handle);
  if (errno < 0) {
    load_entry.size_hi = 0xFF;
    load_entry.size_lo = 0xFF;
    errorf("seeking in index: %ld\n", errno);
  }
  int nRead = read(index_file_handle, &load_entry, sizeof(resource_entry_t));
  if (nRead != sizeof(resource_entry_t)) {
    errorf("reading from index: %d\n", nRead);
  }
  return load_entry;
}

#define NUM_RESOURCE_BANKS 22
typedef struct {
  unsigned char type;
  unsigned char num;
} resource_bank_slot_t;
resource_bank_slot_t resource_bank_tenants[NUM_RESOURCE_BANKS];
unsigned char resource_banks_used = 0;

char *resourceNames[]={"LOGIC", "PIC", "VIEW", "SOUND"};

unsigned char load_resource(unsigned char type, unsigned char num){
  if (resource_banks_used >= NUM_RESOURCE_BANKS){
    errorf("No more resource banks available\n");
    return 0xFF;
  }
  unsigned char thisBank = resource_banks_used++;
  infof("Loading %s %d into bank %d\n", resourceNames[type], num, thisBank);
  getResourceIndex(type, num);
  unsigned long offset = resource_offset(load_entry);
  unsigned int size = resource_size(load_entry);
  push_bank(thisBank);
  int start = clock();
  long errno = f_lseek(offset, SEEK_SET, data_file_handle);
  if (errno < 0) {
    errorf("seeking in data: %ld\n", errno);
  }
  int nRead = read(data_file_handle, RESOURCE_ADDR, size);
  if (nRead != size) {
    errorf("reading from data: %d\n", nRead);
  }
  infof("Loaded %04x bytes %x %x %x %d ms \n", size, RESOURCE_ADDR[0], RESOURCE_ADDR[1], RESOURCE_ADDR[2], clock() - start);
  pop_bank();
  return thisBank;
}