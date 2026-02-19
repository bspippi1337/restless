# Restless C Core (skeleton) — no external deps
# Works on Debian/Termux. For Windows/macOS use CMake later.

CC ?= cc
CFLAGS ?= -O2 -Wall -Wextra -Wpedantic -std=c11
LDFLAGS ?=

BIN_DIR := bin
OBJ_DIR := build
SRC_DIR := src
INC_DIR := include

TARGET := $(BIN_DIR)/restless-core

SRCS := $(SRC_DIR)/main.c $(SRC_DIR)/util.c $(SRC_DIR)/doctor.c $(SRC_DIR)/discovery.c
OBJS := $(SRCS:$(SRC_DIR)/%.c=$(OBJ_DIR)/%.o)

.PHONY: all clean run doctor discover help

all: $(TARGET)

$(TARGET): $(OBJS)
	@mkdir -p $(BIN_DIR)
	$(CC) $(CFLAGS) -I$(INC_DIR) -o $@ $^ $(LDFLAGS)

$(OBJ_DIR)/%.o: $(SRC_DIR)/%.c
	@mkdir -p $(OBJ_DIR)
	$(CC) $(CFLAGS) -I$(INC_DIR) -c -o $@ $<

clean:
	rm -rf $(BIN_DIR) $(OBJ_DIR)

run: all
	$(TARGET)

doctor: all
	$(TARGET) doctor

discover: all
	$(TARGET) discover openai.com

help: all
	$(TARGET) --help
