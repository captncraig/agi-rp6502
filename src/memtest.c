#include <rp6502.h>
#include <stdio.h>

/* Full-coverage sweep, to confirm the whole window after the A6 repair.
 *
 * Rules learned debugging this:
 *  - Stay below $7F00; the soft stack grows down from $8000 and printf digs
 *    deep into it, so anything above that is the program's own scratch.
 *  - Never printf from inside the verify loop; that clobbers the stack and
 *    cascades one real mismatch into dozens of fake ones.
 *  - Patterns must include a low-byte one. A high-byte-only pattern cannot
 *    see single address-line aliasing within a page.
 */

#define LO_START 0x5000
#define LO_END   0x7F00
#define HI_START 0x8000
#define HI_END   0xDD00

static unsigned long fails;
static unsigned first_fail;
static unsigned char first_want, first_got;

static void note(unsigned addr, unsigned char want, unsigned char got)
{
    if (!fails) {
        first_fail = addr;
        first_want = want;
        first_got = got;
    }
    fails++;
}

static unsigned char pat(unsigned addr, unsigned char mode)
{
    switch (mode) {
    case 0: return (unsigned char)addr;
    case 1: return ~(unsigned char)addr;
    case 2: return (unsigned char)(addr >> 8);
    default: return 0x00;
    }
}

static void sweep(unsigned start, unsigned end, unsigned char mode)
{
    unsigned addr;

    for (addr = start; addr < end; addr++)
        *(volatile unsigned char *)addr = pat(addr, mode);
    for (addr = start; addr < end; addr++) {
        unsigned char want = pat(addr, mode);
        unsigned char got = *(volatile unsigned char *)addr;
        if (got != want)
            note(addr, want, got);
    }
    printf(".");
}

int main(void)
{
    unsigned char mode, bank;

    printf("Low RAM + above-window, VIA untouched\n");
    for (mode = 0; mode < 3; mode++) {
        sweep(LO_START, LO_END, mode);
        sweep(0xC000, HI_END, mode);
    }
    printf("\n");

    VIA.ddrb = 0xff;
    printf("Banked window, all 64 banks\n");
    for (bank = 0; bank < 64; bank++) {
        VIA.prb = bank;
        for (mode = 0; mode < 3; mode++)
            sweep(HI_START, 0xC000, mode);
    }
    printf("\n");

    if (fails)
        printf("FAIL: %lu mismatches, first $%04X wrote %02X read %02X\n",
               fails, first_fail, first_want, first_got);
    else
        printf("PASS: no mismatches\n");

    return 0;
}
