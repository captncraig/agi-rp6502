#include "logic.h"
#include <stdint.h>

// stack of loaded programs
typedef struct {
  uint8_t num;
  uint32_t offset;
} logic_entry_t;

#define MAX_LOGIC_DEPTH 5
logic_entry_t logic_stack[MAX_LOGIC_DEPTH];
uint8_t logics_loaded;

void load_logic(uint8_t num){
    logic_stack[logics_loaded].num = num;
}