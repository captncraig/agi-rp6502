#include <rp6502.h>
#include <stdio.h>
#include <fcntl.h>

#include "resource_data.h"
#include "error.h"
#include "vga.h"
#include "pics/pic.h"

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

    init_resource_data();
    init_video();
    clear_screen();
    draw_pic(80);
    show_pic();


    while(1){
        // do stuff that could fatalf at any time.
    }
    
    return 0;
}
    