#include <rp6502.h>
#include <stdio.h>
#include <fcntl.h>

#include "resource_data.h"
#include "error.h"
#include "vga.h"
#include "pics/pic.h"
#include "sound/sound.h"

int main(void)
{
    init_resource_data();
    init_video();
    while(1){
        for (int i = 1; i<100; i++){
            clear_screen();
            if (draw_pic(i) != -1){show_pic();}
        }
    }
    while(1){
        // do stuff that could fatalf at any time.
    }
    return 0;
}
    