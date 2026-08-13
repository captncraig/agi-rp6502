#ifndef _VGA_H
#define _VGA_H

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