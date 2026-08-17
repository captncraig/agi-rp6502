/* Picocomputer 6502 (RP6502) linker script - banked-memory variant.
 *
 * Same output format/pipeline as the stock rp6502 link.ld (one flat image,
 * consumed by tools/rp6502.py's single-chunk loader), but splits the address
 * space into two regions instead of one:
 *
 *   code (0x0200-0x7FFF): .text, .rodata, and .data's LMA (initializer bytes)
 *   data (0xC000-0xFEFF): .data's VMA, .bss, .noinit -- all NOLOAD or
 *                          initialized-in-place at runtime by the standard
 *                          crt0 (copy-data.c / zero-bss.c), so NONE of this
 *                          region's bytes need to live in the shipped image.
 *
 * 0x8000-0xBFFF is deliberately left out of both regions, reserved for a
 * future bank-switched window. Because the "data" region only ever holds
 * NOLOAD/copied-in-place sections, TRIM() only has to cover "code" -- the
 * output image is unaffected by the gap or by how much of "data" is used,
 * so this drops in with zero changes to rp6502_executable/rp6502.py.
 */

/* Provide imaginary (zero page) registers. */
__rc0 = 0x00;
INCLUDE imag-regs.ld
ASSERT(__rc31 == 0x001f, "Inconsistent zero page map.")

MEMORY {
    zp : ORIGIN = __rc31 + 1, LENGTH = 0x100 - (__rc31 + 1)
    code (rx)  : ORIGIN = 0x0200, LENGTH = 0x7E00 /* 0x0200-0x7FFF */
    data (rw)  : ORIGIN = 0xC000, LENGTH = 0x3F00 /* 0xC000-0xFEFF */
}

REGION_ALIAS("c_readonly", code)
REGION_ALIAS("c_writeable", data)

SECTIONS { INCLUDE c.ld }

/* Set initial soft stack address to just above last memory address. (It grows down.) */
__stack = 0xFF00;

OUTPUT_FORMAT {
    SHORT(0x0200)
    SHORT(_start)
    TRIM(code)
}
