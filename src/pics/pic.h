#ifndef _PIC_H
#define _PIC_H

#include <stdint.h>

// clear screen buffers to default state
void clear_screen();
// draw pic over whatever is in buffers
void draw_pic(uint8_t num);
// copy visual buffer to vram
void show_pic();

#endif