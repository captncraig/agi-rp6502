#include "resource_data.h"
#include "error.h"
#include <fcntl.h>
#include <rp6502.h>
#include <stdio.h>
#include <stdint.h>
#include <unistd.h>
#include <time.h>

int data_file_handle = -1;

int init_resource_data(){
  data_file_handle = open("ROM:data.bin", O_RDONLY);
  if (data_file_handle < 0){
    errorf("opening ROM:data.bin: %d\n", data_file_handle);
    return data_file_handle;
  }
  infof("Opened resource data with handle %d\n", data_file_handle);
  return 0;
}

void seek_resource(unsigned long addr){
  if (f_lseek(addr, SEEK_SET, data_file_handle) == -1){
    fatalf("seek error %lx", addr);
  }
}

int data_file(){
  return data_file_handle;
}