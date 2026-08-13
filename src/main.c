#include <rp6502.h>
#include <stdio.h>
#include <fcntl.h>

#include "resource_data.h"
#include "banking.h"
#include "error.h"
#include "vga.h"

// AGI Memory layout:
// 0x0000-0x7fff: C managed code and data
// 0x8000-0xbfff: Banked memory
//     Bank 0: 
//     Bank 1: Graphics Screen (160x200, 4bpp)
//     Bank 2: Priority Screen (160x200, 4bpp)
//     Banks 3-63: Dynamically loaded game resources

void crashout(){
    // change video back to console, stop event loop, whatever.
}

int main(void)
{
    if (setjmp(fatal_jmp_buf)) {
        crashout();
        return 1;
    }

    init_banks();
    if (init_resource_data()){
        printf("!!\n");
        return 1;
    }
    unsigned char bank = load_resource(RESOURCE_TYPE_LOGIC, 0);
    change_bank(bank);
    infof("Loaded %d %x %x %x \n", bank, RESOURCE_ADDR[0], RESOURCE_ADDR[1], RESOURCE_ADDR[2]);
    init_video();

    while(1){
        // do stuff that could fatalf at any time.
    }
    
    return 0;
}
    