#include "pic.h"
#include "../error.h"
#include "../resource_data.h"
#include "../vga.h"
#include <stdint.h>
#include <string.h>
#include <stdbool.h>
#include <rp6502.h>
#include <time.h>


// Visual/priority planes live in XRAM (VIZ_VRAM_START/PRI_VRAM_START), not
// main RAM, freeing ~13KB. addr0/step0/rw0 always drives the viz plane and
// addr1/step1/rw1 the pri plane; the two channels are independent, so
// lockstep sequential passes (clear, span fill, blit) can drive both without
// ever re-issuing an address for either.
#define SCREEN_BYTES (160UL*168/2)

void clear_screen(){
    RIA.step0 = 1;
    RIA.addr0 = VIZ_VRAM_START;
    RIA.step1 = 1;
    RIA.addr1 = PRI_VRAM_START;
    for (uint16_t i = 0; i < SCREEN_BYTES; i++){
        RIA.rw0 = 0xFF;   // color 15 in both nibbles
        RIA.rw1 = 0x44;   // color 4 in both nibbles
    }
}

static uint16_t n_loaded;
static uint16_t pic_size;
static uint16_t chunk_pos;
uint8_t vis_color;
uint8_t pri_color;
bool vis_on;
bool pri_on;
static uint8_t pat_code; // brush shape/size/splatter bits for opcode 0xfa
static uint8_t pat_num;  // splatter PRNG seed, only read when pat_code & 0x20

// addr0/step0 also gets driven by set_pixel/write_span_plane/the viz cursors
// in between calls, so this can't lean on auto-increment surviving across
// calls like it used to — every call re-points addr0 at its own tracked
// offset before reading.
static uint8_t get_next(){
    if (!n_loaded){
        uint16_t chunk = pic_size < PIC_LOAD_SIZE ? pic_size : PIC_LOAD_SIZE;
        int n = read_xram(PIC_VRAM_START, chunk, data_file());
        //infof("Loaded %d bytes for pic\n", n);
        n_loaded = chunk;
        chunk_pos = 0;
    }
    RIA.step0 = 1;
    RIA.addr0 = PIC_VRAM_START + chunk_pos;
    uint8_t b = RIA.rw0;
    chunk_pos++;
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

// Random-access single-pixel read-modify-write. Setting step to 0 makes the
// address register stay put across the read and the write, so this costs one
// addr write + one read + one write per plane, same as a local-RAM RMW would
// cost register-wise, just over the XRAM bus instead.
void set_pixel(uint8_t x1,uint8_t y1){
    uint16_t idx = row_start[y1] + (x1 >> 1);
    if (vis_on){
        RIA.step0 = 0;
        RIA.addr0 = VIZ_VRAM_START + idx;
        uint8_t old = RIA.rw0;
        RIA.rw0 = (x1 & 1) ? ((old & 0xf0) | (vis_color & 0x0f))
                            : ((old & 0x0f) | (vis_color << 4));
    }
    if (pri_on){
        RIA.step1 = 0;
        RIA.addr1 = PRI_VRAM_START + idx;
        uint8_t old = RIA.rw1;
        RIA.rw1 = (x1 & 1) ? ((old & 0xf0) | (pri_color & 0x0f))
                            : ((old & 0x0f) | (pri_color << 4));
    }
}

// Which plane decides fillability for the flood fill currently in progress.
// Matches AGI's real draw_FillCheck rule: with only priority on, priority
// decides; otherwise (visual on, alone or with priority) visual decides and
// priority is never consulted, even though it still gets painted. Set once
// per flood_fill() call from vis_on/pri_on, which don't change mid-fill.
static bool fill_use_pri;

// A pixel is fillable if the plane that governs this fill (see fill_use_pri)
// is still at its background sentinel (0x0f visual / 0x04 priority).
static bool pixel_fillable(uint8_t x,uint8_t y){
    uint16_t idx = row_start[y] + (x >> 1);
    if (fill_use_pri){
        RIA.step1 = 0;
        RIA.addr1 = PRI_VRAM_START + idx;
        uint8_t pb = RIA.rw1;
        uint8_t pri = (x & 1) ? (pb & 0x0f) : (pb >> 4);
        return pri == 0x04;
    }
    RIA.step0 = 0;
    RIA.addr0 = VIZ_VRAM_START + idx;
    uint8_t vb = RIA.rw0;
    uint8_t vis = (x & 1) ? (vb & 0x0f) : (vb >> 4);
    return vis == 0x0f;
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

// Circular brush masks, one row-set per pen size (0-7). circle_list[size]
// gives the starting index into circle_data for that size's rows; each row
// is a bitmask (read via binary_list) of which columns are "inside" the
// circle at that row. Ported from ScummVM's AGI PictureMgr (data table
// itself credited there to NAGI).
static const uint16_t circle_data[] = {
    0x8000,
    0x0000, 0xE000, 0x0000,
    0x7000, 0xF800, 0xF800, 0xF800, 0x7000,
    0x3800, 0x7C00, 0xFE00, 0xFE00, 0xFE00, 0x7C00, 0x3800,
    0x1C00, 0x7F00, 0xFF80, 0xFF80, 0xFF80, 0xFF80, 0xFF80, 0x7F00, 0x1C00,
    0x0E00, 0x3F80, 0x7FC0, 0x7FC0, 0xFFE0, 0xFFE0, 0xFFE0, 0x7FC0, 0x7FC0, 0x3F80, 0x1F00, 0x0E00,
    0x0F80, 0x3FE0, 0x7FF0, 0x7FF0, 0xFFF8, 0xFFF8, 0xFFF8, 0xFFF8, 0xFFF8, 0x7FF0, 0x7FF0, 0x3FE0, 0x0F80,
    0x07C0, 0x1FF0, 0x3FF8, 0x7FFC, 0x7FFC, 0xFFFE, 0xFFFE, 0xFFFE, 0xFFFE, 0xFFFE, 0x7FFC, 0x7FFC, 0x3FF8, 0x1FF0, 0x07C0,
};
static const uint8_t circle_list[] = {0, 1, 4, 9, 16, 25, 37, 50};
static const uint16_t binary_list[] = {
    0x8000, 0x4000, 0x2000, 0x1000, 0x0800, 0x0400, 0x0200, 0x0100,
    0x0080, 0x0040, 0x0020, 0x0010, 0x0008, 0x0004, 0x0002, 0x0001,
};

// Stamps the current brush (pat_code: bits 0-2 size, bit 4 circle vs square,
// bit 5 splatter/dither) centered on (x,y). For each row of the shape we walk
// across pen_width columns 4 at a time; a column is "inside" the brush if
// it's a square (bit 4 set, always inside) or its bit is set in that row's
// circle mask. When splatter is on, an 8-bit LFSR-ish sequence (seeded from
// pat_num) additionally thins those pixels down to ~1/4 of them, giving the
// speckled/dithered look AGI uses for things like grass and trees. Ported
// from ScummVM's AGI PictureMgr::plotPattern.
static void plot_pattern(uint8_t x, uint8_t y){
    uint16_t pen_size = pat_code & 0x07;
    const uint16_t *circle_ptr = &circle_data[circle_list[pen_size]];

    int16_t pen_x = (int16_t)x * 2 - pen_size;
    if (pen_x < 0) pen_x = 0;
    int16_t max_x = 160 * 2 - 2 * pen_size;
    if (pen_x >= max_x) pen_x = max_x;
    pen_x /= 2;
    int16_t pen_start_x = pen_x;

    int16_t pen_y = (int16_t)y - pen_size;
    if (pen_y < 0) pen_y = 0;
    int16_t max_y = 167 - 2 * pen_size;
    if (pen_y >= max_y) pen_y = max_y;

    uint16_t height = (pen_size << 1) + 1;
    int16_t pen_final_y = pen_y + height;
    uint16_t pen_width = height << 1;

    bool is_circle = (pat_code & 0x10) != 0;
    bool is_splatter = (pat_code & 0x20) != 0;
    uint8_t t = pat_num | 0x01;

    for (; pen_y < pen_final_y; pen_y++){
        uint16_t circle_word = *circle_ptr++;
        for (uint16_t counter = 0; counter <= pen_width; counter += 4){
            if (is_circle || (binary_list[counter >> 1] & circle_word)){
                if (is_splatter){
                    uint8_t bit = t & 1;
                    t >>= 1;
                    if (bit) t ^= 0xB8;
                }
                if (!is_splatter || (t & 0x03) == 0x02){
                    set_pixel((uint8_t)pen_x, (uint8_t)pen_y);
                }
            }
            pen_x++;
        }
        pen_x = pen_start_x;
    }
}

// Span-based (scanline) fill: the stack holds one seed pixel per contiguous
// empty run discovered on a row, not one entry per pixel, so it needs far
// fewer slots than the old pixel-at-a-time stack.
#define FILL_STACK_SIZE 2048
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

// Writes color across pixels [lx,rx] on one plane. Full bytes inside the
// range are written directly (no read needed); only the two boundary bytes
// may need a read-modify-write to preserve the nibble outside the range.
static void write_span_plane(volatile unsigned int *addr_reg, volatile unsigned char *step_reg, volatile unsigned char *rw_reg,
                              uint16_t base, uint16_t row_off, uint8_t lx, uint8_t rx, uint8_t color){
    uint16_t first_byte = row_off + (lx >> 1);
    uint16_t last_byte = row_off + (rx >> 1);
    uint8_t byte_color = (color << 4) | (color & 0x0f);

    *step_reg = 0;

    if (first_byte == last_byte){
        *addr_reg = base + first_byte;
        uint8_t old = *rw_reg;
        uint8_t val = old;
        if (!(lx & 1)) val = (val & 0x0f) | (color << 4); // high nibble in range
        if (rx & 1)    val = (val & 0xf0) | color;         // low nibble in range
        *rw_reg = val;
        return;
    }

    uint16_t full_start = first_byte;
    if (lx & 1){
        *addr_reg = base + first_byte;
        uint8_t old = *rw_reg;
        *rw_reg = (old & 0xf0) | color;
        full_start = first_byte + 1;
    }

    uint16_t full_end = last_byte;
    if (!(rx & 1)){
        *addr_reg = base + last_byte;
        uint8_t old = *rw_reg;
        *rw_reg = (old & 0x0f) | (color << 4);
        full_end = last_byte - 1;
    }

    if (full_start <= full_end){
        *step_reg = 1;
        *addr_reg = base + full_start;
        for (uint16_t i = full_start; i <= full_end; i++){
            *rw_reg = byte_color;
        }
    }
}

// Sequential left-to-right reader over the plane fill_use_pri selects: the
// address register only moves once every two pixels (one XRAM byte packs
// two nibbles), instead of re-addressing on every single pixel.
typedef struct { uint16_t idx; bool odd; uint8_t byte; } plane_cursor_t;

static plane_cursor_t fill_cursor_begin(uint8_t x, uint8_t y){
    plane_cursor_t c;
    c.idx = row_start[y] + (x >> 1);
    c.odd = x & 1;
    if (fill_use_pri){
        RIA.step1 = 0;
        RIA.addr1 = PRI_VRAM_START + c.idx;
        c.byte = RIA.rw1;
    } else {
        RIA.step0 = 0;
        RIA.addr0 = VIZ_VRAM_START + c.idx;
        c.byte = RIA.rw0;
    }
    return c;
}
static uint8_t fill_cursor_next(plane_cursor_t *c){
    uint8_t nib = c->odd ? (c->byte & 0x0f) : (c->byte >> 4);
    if (c->odd){
        c->idx++;
        if (fill_use_pri){
            RIA.addr1 = PRI_VRAM_START + c->idx;
            c->byte = RIA.rw1;
        } else {
            RIA.addr0 = VIZ_VRAM_START + c->idx;
            c->byte = RIA.rw0;
        }
    }
    c->odd = !c->odd;
    return nib;
}

// Scans row y across [lx,rx], pushing one seed per contiguous fillable run.
// Streams the governing plane (fill_use_pri) sequentially instead of calling
// pixel_fillable per pixel. When a whole cached byte is the packed
// background sentinel (0xFF viz / 0x44 pri), both its pixels are fillable
// with no nibble extraction needed, so the pair is skipped in one step.
static void fill_scan(uint8_t lx, uint8_t rx, uint8_t y, uint16_t *count){
    uint8_t sentinel_byte = fill_use_pri ? 0x44 : 0xFF;
    uint8_t sentinel_nib = fill_use_pri ? 0x04 : 0x0f;
    uint16_t base = fill_use_pri ? PRI_VRAM_START : VIZ_VRAM_START;
    bool added = false;
    plane_cursor_t c = fill_cursor_begin(lx, y);
    uint8_t x = lx;
    while (x <= rx){
        if (!(x & 1) && x + 1 <= rx && c.byte == sentinel_byte){
            if (!added){
                fill_push(count, x, y);
                added = true;
            }
            c.idx++;
            if (fill_use_pri){ RIA.addr1 = base + c.idx; c.byte = RIA.rw1; }
            else { RIA.addr0 = base + c.idx; c.byte = RIA.rw0; }
            x += 2;
            continue;
        }
        if (fill_cursor_next(&c) != sentinel_nib){
            added = false;
        } else if (!added){
            fill_push(count, x, y);
            added = true;
        }
        x++;
    }
}

uint32_t flood_fill(uint8_t x,uint8_t y){
    if (!vis_on && !pri_on){
        return 0;
    }
    // Matches AGI's real fill-check: priority only governs when it's the
    // sole active plane; otherwise visual governs (and gets tested), even
    // with priority also on. A fill color equal to its plane's background
    // sentinel is defined as a no-op, same as the original engine.
    fill_use_pri = pri_on && !vis_on;
    if (fill_use_pri){
        if (pri_color == 0x04) return 0;
    } else {
        if (vis_color == 0x0f) return 0;
    }

    uint16_t stack_count = 0;
    uint32_t fills = 0;
    if (!pixel_fillable(x,y)){
        return 0;
    }
    fill_push(&stack_count, x, y);
    while(stack_count){
        stack_count--;
        x = fill_stack_x[stack_count];
        y = fill_stack_y[stack_count];
        if (!pixel_fillable(x,y)){
            continue; // already covered by an earlier span this round
        }

        uint8_t lx = x;
        while (lx > 0 && pixel_fillable(lx-1,y)) lx--;
        uint8_t rx = x;
        while (rx < 159 && pixel_fillable(rx+1,y)) rx++;

        if (vis_on) write_span_plane(&RIA.addr0, &RIA.step0, &RIA.rw0, VIZ_VRAM_START, row_start[y], lx, rx, vis_color);
        if (pri_on) write_span_plane(&RIA.addr1, &RIA.step1, &RIA.rw1, PRI_VRAM_START, row_start[y], lx, rx, pri_color);
        fills += (uint32_t)(rx - lx + 1);

        if (y > 0) fill_scan(lx, rx, y-1, &stack_count);
        if (y < 167) fill_scan(lx, rx, y+1, &stack_count);
    }
    return fills;
}

int draw_pic(uint8_t num){
    if (!row_table_ready){
        init_row_table();
    }
    resource_entry_t entry = resource_index.pics[num];
    if(!RESOURCE_PRESENT(entry)){
        errorf("Pic no exist %d\n", num);
        return -1;
    }
    seek_resource(entry.offset);
    pic_size = entry.size;
    n_loaded = 0;
    fill_stack_watermark = 0;
    RIA.step0 = 1;
    bool peeked = false;
    uint8_t peek;
    vis_color = 0x0f;
    pri_color = 0x04;
    vis_on = false;
    pri_on = false;
    pat_code = 0;
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
                //infof("Viz color %x\n", vis_color);
                break;
            case 0xf1:
                vis_on = false;
                //infof("Viz off\n");
                break;
            case 0xf2:
                pri_on = true;
                pri_color = get_next() & 0x0f;
                //infof("Pri color %x\n", pri_color);
                break;
            case 0xf3:
                pri_on = false;
                //infof("Pri off\n");
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
                    //infof("Line %d,%d -> %d,%d\n", x1,y1,x2,y2);
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
                    //infof("Short line %d,%d -> %d,%d\n", x1,y1,x2,y2);
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
                    //infof("Flood %d,%d ", x1,y1);
                    flood_fill(x1,y1);
                    //infof("(%ld)\n", flood_fill(x1,y1));
                }
                break;
            case 0xf9:
                pat_code = get_next();
                break;
            case 0xfa:
                while(1){
                    if (pat_code & 0x20){
                        pat_num = get_next();
                        if (pat_num >= 0xf0){
                            peeked = true;
                            peek = pat_num;
                            break;
                        }
                    }
                    x1 = get_next();
                    if (x1 >= 0xf0){
                        peeked = true;
                        peek = x1;
                        break;
                    }
                    y1 = get_next();
                    plot_pattern(x1,y1);
                }
                break;
            default:
                infof("Pic opcode %x\n", op);
                //infof("Fill stack watermark: %u, draw time: %lu ms\n", fill_stack_watermark, (unsigned long)(clock() - start));
                return 0;
        }
    }
    infof("Pic %d done in %lu ms, fill stack watermark %u\n", num, (unsigned long)(clock() - start), fill_stack_watermark);
    return 0;
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
    RIA.addr1 = VIZ_VRAM_START;
    RIA.step1 = 1;
    for (uint16_t i = 0; i < SCREEN_BYTES; i++){
        uint8_t pix = RIA.rw1;
        RIA.rw0 = nibble_hi[pix];
        RIA.rw0 = nibble_lo[pix];
    }
}