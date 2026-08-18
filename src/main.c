#include <rp6502.h>
#include <6502.h>
#include <stdio.h>
#include <fcntl.h>

#include "resource_data.h"
#include "error.h"
#include "vga.h"
#include "pics/pic.h"
#include "sound/sound.h"

#define RIA_IRQ_VSYNC 0x80

static volatile unsigned char vsync_flag;

__attribute__((interrupt)) void irq_isr(void)
{
    if (RIA.irq & RIA_IRQ_VSYNC)
        vsync_flag = 1;
}

static void wait_vsync(void)
{
    while (!vsync_flag) {}
    vsync_flag = 0;
}

int main(void)
{
    init_resource_data();
    init_video();
    RIA.irq = RIA_IRQ_VSYNC;
    CLI();
    while(1){
        for (int i = 1; i<100; i++){
            wait_vsync();
            clear_screen();
            if (draw_pic(i) != -1){show_pic();}
        }
    }
    return 0;
}
    