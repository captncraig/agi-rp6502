#include <rp6502.h>
#include <stdio.h>
#include <fcntl.h>

#include "resource_data.h"
#include "banking.h"
#include "error.h"

// AGI Memory layout:
// 0x0000-0x7fff: C managed code and data
// 0x8000-0xbfff: Banked memory
//     Bank 0: 
//     Bank 1: Graphics Screen (160x200, 4bpp)
//     Bank 2: Priority Screen (160x200, 4bpp)
//     Banks 3-63: Dynamically loaded game resources

int main(void)
{
    init_banks();
    if (init_resource_data()){
        printf("!!\n");
        return 1;
    }

    unsigned char bank = load_resource(RESOURCE_TYPE_LOGIC, 0);
    change_bank(bank);
    infof("Loaded %d %x %x %x \n", bank, RESOURCE_ADDR[0], RESOURCE_ADDR[1], RESOURCE_ADDR[2]);
    return 0;
}
    