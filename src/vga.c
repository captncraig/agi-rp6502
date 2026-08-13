#include "vga.h"
#include <rp6502.h>
#include <stdio.h>
#include <stdbool.h>



// vram layout:
// 1:0000-1:68FF 320x168 4bpp bitmap
// 1:6900-1:70C9 Text layer (40x25)
// 1:70D0 - config structs
// 1:70EE - palette
// 1:710E-1:FFFF Free scratch xram!
#define BITMAP_SIZE (320L*168/2)
#define BG_VRAM_START (0L)
#define TEXT_LAYER_SIZE (40L*25*2)
#define TEXT_VRAM_START (BG_VRAM_START + BITMAP_SIZE)
#define CONFIG0_VRAM_START (TEXT_VRAM_START + TEXT_LAYER_SIZE)
#define CONFIG1_VRAM_START (CONFIG0_VRAM_START + sizeof(vga_mode3_config_t))
#define PALETTE_VRAM_START (CONFIG1_VRAM_START + sizeof(vga_mode1_config_t))
#define XRAM_FREE (PALETTE_VRAM_START + 32)

static void write_palette(unsigned int xram_addr, const unsigned int *colors, int count);

void init_video(){
    printf("BG VRAM start: %lx\n", BG_VRAM_START);
    printf("TEXT VRAM start: %lx\n", TEXT_VRAM_START);
    printf("Config0 VRAM start: %lx\n", CONFIG0_VRAM_START);
    printf("Config1 VRAM start: %lx\n", CONFIG1_VRAM_START);
    printf("Palette VRAM start: %lx\n", PALETTE_VRAM_START);
    printf("FREE XRAM start: %lx\n", XRAM_FREE);

    // 320x240
    xreg_vga_canvas(1);

    // set up background
    xram0_struct_set(CONFIG0_VRAM_START, vga_mode3_config_t, x_wrap, false);
    xram0_struct_set(CONFIG0_VRAM_START, vga_mode3_config_t, y_wrap, false);
    xram0_struct_set(CONFIG0_VRAM_START, vga_mode3_config_t, x_pos_px, 0);
    xram0_struct_set(CONFIG0_VRAM_START, vga_mode3_config_t, y_pos_px, 8);
    xram0_struct_set(CONFIG0_VRAM_START, vga_mode3_config_t, width_px, 320);
    xram0_struct_set(CONFIG0_VRAM_START, vga_mode3_config_t, height_px, 168);
    xram0_struct_set(CONFIG0_VRAM_START, vga_mode3_config_t, xram_data_ptr, BG_VRAM_START);
    xram0_struct_set(CONFIG0_VRAM_START, vga_mode3_config_t, xram_palette_ptr, 0xffff);

    xreg_vga_mode(3, 2, (unsigned)CONFIG0_VRAM_START, 0, 8, 176);

    write_palette(PALETTE_VRAM_START, agi_palette, 16);

    RIA.addr0 = BG_VRAM_START;
    RIA.step0 = 1;
    for (unsigned char bar = 0; bar < 16; bar++){
        unsigned char colorByte = bar | (bar << 4);
        for (unsigned char row = 0; row < 10; row++){
            for (unsigned char x = 0; x < 160; x++){
                RIA.rw0 = colorByte;
            }
        }
    }
}

static void write_palette(unsigned int xram_addr, const unsigned int *colors, int count){
    RIA.addr0 = xram_addr;
    RIA.step0 = 1;
    for (int i = 0; i < count; i++){
        RIA.rw0 = colors[i] & 0xff;
        RIA.rw0 = (colors[i] >> 8) & 0xff;
    }
}