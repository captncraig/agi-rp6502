cmake_minimum_required(VERSION 3.18)

# Default to -O3 unless the caller explicitly asks for a Debug/RelWithDebInfo
# build; without this an unconfigured build dir silently lands on CMake's
# own default of Debug (no -O flag at all).
if(NOT CMAKE_BUILD_TYPE)
    set(CMAKE_BUILD_TYPE Release CACHE STRING "" FORCE)
endif()

add_subdirectory(tools)

set(LLVM_MOS_PLATFORM rp6502)
# Prefer the SDK built from the vendored llvm-mos-sdk submodule (kept newer than
# any globally installed llvm-mos release) over anything else on the search path.
list(PREPEND CMAKE_PREFIX_PATH "${CMAKE_CURRENT_SOURCE_DIR}/llvm-mos-sdk/install")
find_package(llvm-mos-sdk REQUIRED)

project(AGI-RP6502 C CXX ASM)

# The compiler's own auto-discovered target config file lives next to the compiler
# binary and resolves via realpath, so it always finds the global ~/llvm-mos install
# regardless of which symlink was invoked. Disable that and point explicitly at the
# submodule's own cfg so headers (e.g. VIA) come from llvm-mos-sdk/install, not ~/llvm-mos.
set(RP6502_LOCAL_CFG "${CMAKE_CURRENT_SOURCE_DIR}/llvm-mos-sdk/install/bin/mos-rp6502.cfg")
add_compile_options(--no-default-config "--config=${RP6502_LOCAL_CFG}")

# Opt-in banked memory layout: code lives at 0x0200-0x7FFF, writable/bss data
# at 0xC000-0xFEFF, leaving 0x8000-0xBFFF unused/reserved for a future
# bank-switched window. See linker/rp6502/link.ld for why this needs no
# changes to the rp6502.py upload pipeline.
#
# The mos linker driver always appends a bare "-Tlink.ld", resolved via
# whichever -L directory containing that filename comes first -- and an -L
# given directly on the command line is always placed by the driver AFTER
# the ones baked into the --config file, so a plain add_link_options(-L...)
# can't win the race. linker/rp6502/mos-rp6502-banked.cfg works around this
# by putting our -L ahead of the SDK's own cfg (via @include), inside the
# config file itself.
option(RP6502_BANKED_LINK "Use linker/rp6502/link.ld instead of the SDK's default flat link.ld" ON)
if(RP6502_BANKED_LINK)
    set(RP6502_LOCAL_LINK_CFG "${CMAKE_CURRENT_SOURCE_DIR}/linker/rp6502/mos-rp6502-banked.cfg")
else()
    set(RP6502_LOCAL_LINK_CFG "${RP6502_LOCAL_CFG}")
endif()

add_link_options(--no-default-config "--config=${RP6502_LOCAL_LINK_CFG}" "-Wl,--print-memory-usage")

add_executable(agi)
rp6502_asset(agi help src/help.txt)
rp6502_executable(agi DATA file RESET file IRQ file)
target_sources(agi PRIVATE
    src/main.c
    src/resource_data.c
    src/error.c
    src/vga.c
    src/logic/logic.c
    src/pics/pic.c
    src/sound/sound.c
)
