#include "pic.h"
#include "../error.h"
#include "../resource_data.h"
#include "../vga.h"
#include <stdint.h>
#include <string.h>
#include <stdbool.h>
#include <rp6502.h>
#include <time.h>


__attribute__((used)) uint8_t visual_screen[160UL*168/2];
__attribute__((used)) uint8_t priority_screen[160UL*168/2];

void clear_screen(){
    memset(visual_screen, 0xFF, sizeof(visual_screen));   // color 15 in both nibbles
    memset(priority_screen, 0x44, sizeof(priority_screen)); // color 4 in both nibbles
}

static uint16_t n_loaded;
static uint16_t pic_size;
uint8_t vis_color;
uint8_t pri_color;
bool vis_on;
bool pri_on;

static uint8_t get_next(){
    if (!n_loaded){
        uint16_t chunk = pic_size < PIC_LOAD_SIZE ? pic_size : PIC_LOAD_SIZE;
        int n = read_xram(PIC_VRAM_START, chunk, data_file());
        infof("Loaded %d bytes for pic\n", n);
        RIA.addr0 = PIC_VRAM_START;
        n_loaded = chunk;
    }
    uint8_t b = RIA.rw0;
    n_loaded--;
    pic_size--;
    return b;
}

static uint16_t row_start[168];
static uint8_t row_table_ready = 0;

static void init_row_table(){
    uint16_t acc = 0;
    for (uint16_t y = 0; y < 168; y++){
        row_start[y] = acc;
        acc += 80;
    }
    row_table_ready = 1;
}

void set_pixel(uint8_t x1,uint8_t y1){
    uint16_t idx = row_start[y1] + (x1 >> 1);
    if (vis_on){
        uint8_t *p = &visual_screen[idx];
        *p = (x1 & 1) ? ((*p & 0xf0) | (vis_color & 0x0f))
                      : ((*p & 0x0f) | (vis_color << 4));
    }
    if (pri_on){
        uint8_t *p = &priority_screen[idx];
        *p = (x1 & 1) ? ((*p & 0xf0) | (pri_color & 0x0f))
                      : ((*p & 0x0f) | (pri_color << 4));
    }
}

// Returns the current vis color in the high nibble, priority color in the low
// nibble. Only used by flood_fill, which needs both at once.
static uint8_t get_pixel(uint8_t x,uint8_t y){
    uint16_t idx = row_start[y] + (x >> 1);
    uint8_t vis = (x & 1) ? (visual_screen[idx] & 0x0f) : (visual_screen[idx] >> 4);
    uint8_t pri = (x & 1) ? (priority_screen[idx] & 0x0f) : (priority_screen[idx] >> 4);
    return (vis << 4) | pri;
}

// A pixel is fillable only if it's still at background color on every
// screen that's currently enabled for the fill.
static bool pixel_empty(uint8_t x,uint8_t y){
    uint8_t color = get_pixel(x,y);
    if (vis_on && (color >> 4) != 0x0f) return false;
    if (pri_on && (color & 0x0f) != 0x04) return false;
    return true;
}

// Integer-only DDA line draw, ported from ScummVM's AGI PictureMgr::draw_Line
// ("A line drawing routine sent by Joshua Neal, modified by Stuart George").
void draw_line(uint8_t x1,uint8_t y1,uint8_t x2,uint8_t y2){
    if (x1 > 159) x1 = 159;
    if (x2 > 159) x2 = 159;
    if (y1 > 167) y1 = 167;
    if (y2 > 167) y2 = 167;

    if (x1 == x2){
        if (y1 > y2){ uint8_t t = y1; y1 = y2; y2 = t; }
        for (; y1 <= y2; y1++) set_pixel(x1,y1);
        return;
    }

    if (y1 == y2){
        if (x1 > x2){ uint8_t t = x1; x1 = x2; x2 = t; }
        for (; x1 <= x2; x1++) set_pixel(x1,y1);
        return;
    }

    int8_t stepX = 1;
    int16_t deltaX = (int16_t)x2 - (int16_t)x1;
    if (deltaX < 0){ stepX = -1; deltaX = -deltaX; }

    int8_t stepY = 1;
    int16_t deltaY = (int16_t)y2 - (int16_t)y1;
    if (deltaY < 0){ stepY = -1; deltaY = -deltaY; }

    uint16_t i, detdelta, errorX, errorY;
    if (deltaY > deltaX){
        i = deltaY;
        detdelta = deltaY;
        errorX = deltaY / 2;
        errorY = 0;
    } else {
        i = deltaX;
        detdelta = deltaX;
        errorX = 0;
        errorY = deltaX / 2;
    }

    uint8_t x = x1;
    uint8_t y = y1;
    set_pixel(x,y);

    do {
        errorY += deltaY;
        if (errorY >= detdelta){
            errorY -= detdelta;
            y += stepY;
        }
        errorX += deltaX;
        if (errorX >= detdelta){
            errorX -= detdelta;
            x += stepX;
        }
        set_pixel(x,y);
        i--;
    } while (i > 0);
}

// Worst-case frontier for a 160x168 screen is bounded by total pixel count,
// but real AGI fills never come close to that. Sized generously with a hard
// overflow guard rather than trusting it never happens.
#define FILL_STACK_SIZE 9000
static uint8_t fill_stack_x[FILL_STACK_SIZE];
static uint8_t fill_stack_y[FILL_STACK_SIZE];

static uint16_t fill_stack_watermark = 0;

static void fill_push(uint16_t *count, uint8_t x, uint8_t y){
    if (*count >= FILL_STACK_SIZE){
        errorf("Flood fill stack overflow\n");
        return;
    }
    fill_stack_x[*count] = x;
    fill_stack_y[*count] = y;
    (*count)++;
    if (*count > fill_stack_watermark){
        fill_stack_watermark = *count;
    }
}

uint32_t flood_fill(uint8_t x,uint8_t y){
    uint16_t stack_count = 0;
    uint32_t fills = 0;
    if (!pixel_empty(x,y)){
        return 0;
    }
    set_pixel(x,y);
    fills++;
    fill_push(&stack_count, x, y);
    while(stack_count){
        stack_count--;
        x = fill_stack_x[stack_count];
        y = fill_stack_y[stack_count];

        if (x > 0 && pixel_empty(x-1,y)){
            set_pixel(x-1,y);
            fills++;
            fill_push(&stack_count, x-1, y);
        }
        if (x < 159 && pixel_empty(x+1,y)){
            set_pixel(x+1,y);
            fills++;
            fill_push(&stack_count, x+1, y);
        }
        if (y > 0 && pixel_empty(x,y-1)){
            set_pixel(x,y-1);
            fills++;
            fill_push(&stack_count, x, y-1);
        }
        if (y < 167 && pixel_empty(x,y+1)){
            set_pixel(x,y+1);
            fills++;
            fill_push(&stack_count, x, y+1);
        }
    };
    return fills;
}

void draw_pic(uint8_t num){
    if (!row_table_ready){
        init_row_table();
    }
    resource_entry_t entry = getResourceIndex(RESOURCE_TYPE_PIC, num);
    if(!RESOURCE_PRESENT(entry)){
        fatalf("Pic no exist %d\n", num);
    }
    seek_resource(resource_offset(entry));
    pic_size = resource_size(entry);
    n_loaded = 0;
    fill_stack_watermark = 0;
    infof("Pic %d ready. %d bytes\n", num, pic_size);
    RIA.step0 = 1;
    bool peeked = false;
    uint8_t peek;
    vis_color = 0x0f;
    pri_color = 0x04;
    vis_on = false;
    pri_on = false;
    uint8_t x1;
    uint8_t y1;
    uint8_t x2;
    uint8_t y2;
    long start = clock(); // rp6502's clock() returns elapsed ms directly
    while(pic_size){
        uint8_t op = peeked ? peek : get_next();
        peeked = false;
        switch(op){
            case 0xf0:
                vis_on = true;
                vis_color = get_next() & 0x0f;
                infof("Viz color %x\n", vis_color);
                break;
            case 0xf1:
                vis_on = false;
                infof("Viz off\n");
                break;
            case 0xf2:
                pri_on = true;
                pri_color = get_next() & 0x0f;
                infof("Pri color %x\n", pri_color);
                break;
            case 0xf3:
                pri_on = false;
                infof("Pri off\n");
                break;
            case 0xf4: // y-corner: vertical step, then horizontal, alternating
                x1 = get_next();
                y1 = get_next();
                set_pixel(x1,y1);

                while(1){
                    y2 = get_next();
                    if (y2 >= 0xf0){
                        peeked = true;
                        peek = y2;
                        break;
                    }
                    draw_line(x1,y1,x1,y2);
                    y1 = y2;

                    x2 = get_next();
                    if (x2 >= 0xf0){
                        peeked = true;
                        peek = x2;
                        break;
                    }
                    draw_line(x1,y1,x2,y1);
                    x1 = x2;
                }
                break;
            case 0xf5: // x-corner: horizontal step, then vertical, alternating
                x1 = get_next();
                y1 = get_next();
                set_pixel(x1,y1);

                while(1){
                    x2 = get_next();
                    if (x2 >= 0xf0){
                        peeked = true;
                        peek = x2;
                        break;
                    }
                    draw_line(x1,y1,x2,y1);
                    x1 = x2;

                    y2 = get_next();
                    if (y2 >= 0xf0){
                        peeked = true;
                        peek = y2;
                        break;
                    }
                    draw_line(x1,y1,x1,y2);
                    y1 = y2;
                }
                break;
            case 0xf6:
                x1 = get_next();
                y1 = get_next();

                while(1){
                    x2 = get_next();
                    if (x2 >= 0xf0){
                        peeked = true;
                        peek = x2;
                        break;
                    }
                    y2 = get_next();
                    draw_line(x1,y1,x2,y2);
                    infof("Line %d,%d -> %d,%d\n", x1,y1,x2,y2);
                    x1 = x2;
                    y1 = y2;
                }
                break;
            case 0xf7: { // relative/short line: signed 4-bit dx,dy packed in one byte
                x1 = get_next();
                y1 = get_next();

                while(1){
                    uint8_t disp = get_next();
                    if (disp >= 0xf0){
                        peeked = true;
                        peek = disp;
                        break;
                    }
                    int8_t dx = (disp >> 4) & 0x0f;
                    int8_t dy = disp & 0x0f;
                    if (dx & 0x08) dx = -(dx & 0x07);
                    if (dy & 0x08) dy = -(dy & 0x07);

                    int16_t nx = (int16_t)x1 + dx;
                    int16_t ny = (int16_t)y1 + dy;
                    if (nx < 0) nx = 0;
                    if (ny < 0) ny = 0;
                    x2 = (uint8_t)nx;
                    y2 = (uint8_t)ny;

                    draw_line(x1,y1,x2,y2);
                    infof("Short line %d,%d -> %d,%d\n", x1,y1,x2,y2);
                    x1 = x2;
                    y1 = y2;
                }
                break;
            }
            case 0xf8:
                while(1){
                    x1 = get_next();
                    if (x1 >= 0xf0){
                        peeked = true;
                        peek = x1;
                        break;
                    }
                    y1 = get_next();
                    infof("Flood %d,%d ", x1,y1);
                    infof("(%ld)\n", flood_fill(x1,y1));
                }
                break;
            default:
                infof("Pic opcode %x\n", op);
                infof("Fill stack watermark: %u, draw time: %lu ms\n", fill_stack_watermark, (unsigned long)(clock() - start));
                return;
        }
    }
    infof("Fill stack watermark: %u, draw time: %lu ms\n", fill_stack_watermark, (unsigned long)(clock() - start));
}

static uint8_t nibble_hi[256];
static uint8_t nibble_lo[256];
static uint8_t nibble_tables_ready = 0;

static void init_nibble_tables(){
    for (uint16_t n = 0; n < 256; n++){
        uint8_t b = (uint8_t)n;
        nibble_hi[n] = (b & 0xf0) | (b >> 4);
        nibble_lo[n] = (b << 4) | (b & 0x0f);
    }
    nibble_tables_ready = 1;
}

void show_pic(){
    if (!nibble_tables_ready){
        init_nibble_tables();
    }
    RIA.addr0 = BG_VRAM_START;
    RIA.step0 = 1;
    uint16_t i = 0;
    for (uint8_t y = 0; y<168; y++){
        for(uint8_t x = 0; x<80; x++){
            uint8_t pix = visual_screen[i++];
            RIA.rw0 = nibble_hi[pix];
            RIA.rw0 = nibble_lo[pix];
        }
    }
}