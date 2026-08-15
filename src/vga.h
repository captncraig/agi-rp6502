#ifndef _VGA_H
#define _VGA_H

#include <stdint.h>
#include <rp6502.h>

#define PIC_LOAD_SIZE (256)

typedef struct __attribute__((packed)) {
    uint8_t bitmap[320UL*168/2];
    uint8_t text[40UL*25*2];
    vga_mode3_config_t config0;
    vga_mode1_config_t config1;
    uint16_t palette[16];
    uint8_t logic_exe[1024]; // currently executing logic scratch
    uint8_t sound_exe[4*40*5]; // active sound buffer. 40 notes each for 4 voices.
    // OPT: we may find loading cels JIT is too slow at render time. If we have xram to spare, 
    // keeping full views loaded dynamically could save read time at render.
    uint8_t view_cel_load[1024UL*3]; // space for one cel to be rendered
    uint8_t pic_load[PIC_LOAD_SIZE];
} xram_layout_t;

#define BG_VRAM_START       offsetof(xram_layout_t, bitmap)
#define TEXT_VRAM_START     offsetof(xram_layout_t, text)
#define CONFIG0_VRAM_START  offsetof(xram_layout_t, config0)
#define CONFIG1_VRAM_START  offsetof(xram_layout_t, config1)
#define PALETTE_VRAM_START  offsetof(xram_layout_t, palette)
#define LOGIC_VRAM_START    offsetof(xram_layout_t, logic_exe)
#define VIEW_VRAM_START     offsetof(xram_layout_t, view_cel_load)
#define SOUND_VRAM_START    offsetof(xram_layout_t, sound_exe)
#define PIC_VRAM_START      offsetof(xram_layout_t, pic_load)

void init_video();

// r,g,b are 6-bit (0-63) channels, matching the EGA table below.
#define COLOR_ALPHA_MASK (1u << 5)
#define COLOR_FROM_RGB6(r, g, b) \
    ((((b) >> 1) << 11) | (((g) >> 1) << 6) | ((r) >> 1))

// Code  Color             R    G    B
// ----- ---------------- ---- ---- ----
//   0   black            0x00 0x00 0x00
//   1   blue             0x00 0x00 0x2A
//   2   green            0x00 0x2A 0x00
//   3   cyan             0x00 0x2A 0x2A
//   4   red              0x2A 0x00 0x00
//   5   magenta          0x2A 0x00 0x2A
//   6   brown            0x2A 0x15 0x00
//   7   light grey       0x2A 0x2A 0x2A
//   8   dark grey        0x15 0x15 0x15
//   9   light blue       0x15 0x15 0x3F
//  10   light green      0x15 0x3F 0x15
//  11   light cyan       0x15 0x3F 0x3F
//  12   light red        0x3F 0x15 0x15
//  13   light magenta    0x3F 0x15 0x3F
//  14   yellow           0x3F 0x3F 0x15
//  15   white            0x3F 0x3F 0x3F
static const unsigned int agi_palette[16] = {
    COLOR_FROM_RGB6(0x00, 0x00, 0x00) | COLOR_ALPHA_MASK,
    COLOR_FROM_RGB6(0x00, 0x00, 0x2A) | COLOR_ALPHA_MASK,
    COLOR_FROM_RGB6(0x00, 0x2A, 0x00) | COLOR_ALPHA_MASK,
    COLOR_FROM_RGB6(0x00, 0x2A, 0x2A) | COLOR_ALPHA_MASK,
    COLOR_FROM_RGB6(0x2A, 0x00, 0x00) | COLOR_ALPHA_MASK,
    COLOR_FROM_RGB6(0x2A, 0x00, 0x2A) | COLOR_ALPHA_MASK,
    COLOR_FROM_RGB6(0x2A, 0x15, 0x00) | COLOR_ALPHA_MASK,
    COLOR_FROM_RGB6(0x2A, 0x2A, 0x2A) | COLOR_ALPHA_MASK,
    COLOR_FROM_RGB6(0x15, 0x15, 0x15) | COLOR_ALPHA_MASK,
    COLOR_FROM_RGB6(0x15, 0x15, 0x3F) | COLOR_ALPHA_MASK,
    COLOR_FROM_RGB6(0x15, 0x3F, 0x15) | COLOR_ALPHA_MASK,
    COLOR_FROM_RGB6(0x15, 0x3F, 0x3F) | COLOR_ALPHA_MASK,
    COLOR_FROM_RGB6(0x3F, 0x15, 0x15) | COLOR_ALPHA_MASK,
    COLOR_FROM_RGB6(0x3F, 0x15, 0x3F) | COLOR_ALPHA_MASK,
    COLOR_FROM_RGB6(0x3F, 0x3F, 0x15) | COLOR_ALPHA_MASK,
    COLOR_FROM_RGB6(0x3F, 0x3F, 0x3F) | COLOR_ALPHA_MASK,
};

#endif