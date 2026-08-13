GAMES := $(notdir $(patsubst %/,%,$(wildcard games/*/)))

.PHONY: bin unpack decomp repack roms $(GAMES)

# Build build/agi.rp6502 from the C source.
bin:
	cmake --build build
	@ls -l build/agi.rp6502

# Extract AGI resources from games/*/ into games/*/unpacked/.
unpack:
	go run ./cmd/unpack

# Decompile unpacked logic scripts into games/*/decomp/*.txt.
decomp:
	go run ./cmd/decomp

# Concatenate unpacked resources into games/*/data.bin and games/*/index.bin.
repack:
	go run ./cmd/repack

# Merge build/agi.rp6502 with each game's index/data into build/games/*.rp6502.
roms:
	go run ./cmd/makeroms

# Single-game convenience target for testing, e.g. `make sqi`: rebuilds the
# ROM, repacks just that game, and merges it into build/games/<name>.rp6502.
$(GAMES): bin
	go run ./cmd/repack $@
	go run ./cmd/makeroms $@

run:
	python3 ./tools/rp6502.py -k craig -d 192.168.1.118:23 run ./build/games/sqi.rp6502