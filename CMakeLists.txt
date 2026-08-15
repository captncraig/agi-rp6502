cmake_minimum_required(VERSION 3.18)

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
add_link_options(--no-default-config "--config=${RP6502_LOCAL_CFG}" "-Wl,--print-memory-usage")

add_executable(agi)
rp6502_asset(agi help src/help.txt)
rp6502_executable(agi DATA file RESET file)
target_sources(agi PRIVATE
    src/main.c
    src/resource_data.c
    src/error.c
    src/vga.c
    src/logic/logic.c
    src/pics/pic.c
)
