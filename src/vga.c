#include "vga.h"
#include <rp6502.h>
#include <stdio.h>
#include <stdbool.h>

static void write_palette(unsigned int xram_addr, const unsigned int *colors, int count);

void init_video(){
    printf("BG VRAM start: %lx\n", BG_VRAM_START);
    printf("TEXT VRAM start: %lx\n", TEXT_VRAM_START);
    printf("Config0 VRAM start: %lx\n", CONFIG0_VRAM_START);
    printf("Config1 VRAM start: %lx\n", CONFIG1_VRAM_START);
    printf("Palette VRAM start: %lx\n", PALETTE_VRAM_START);
    printf("FREE XRAM: %lx\n",0x10000- sizeof(xram_layout_t));
   // return;
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
    xram0_struct_set(CONFIG0_VRAM_START, vga_mode3_config_t, xram_palette_ptr, PALETTE_VRAM_START);

    xreg_vga_mode(3, 2, (unsigned)CONFIG0_VRAM_START, 0, 8, 176);
    xreg_vga_mode(0,1 , 176,0);
    write_palette(PALETTE_VRAM_START, agi_palette, 16);
}

static void write_palette(unsigned int xram_addr, const unsigned int *colors, int count){
    RIA.addr0 = xram_addr;
    RIA.step0 = 1;
    for (int i = 0; i < count; i++){
        RIA.rw0 = colors[i] & 0xff;
        RIA.rw0 = (colors[i] >> 8) & 0xff;
    }
}