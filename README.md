# dotman

A simple and efficient dotfiles manager written in Go.

## Overview

dotman is a command-line tool designed to help you manage your dotfiles across different machines. It provides a simple way to track, version, and deploy your configuration files.

## Features

- Simple and intuitive command-line interface
- Verbose mode for detailed output
- Easy to build and run
- Cross-platform support

## Installation

### Prerequisites

- Go 1.24 or later
- Make

### Building from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/noosxe/dotman.git
   cd dotman
   ```

2. Build the application:
   ```bash
   make build
   ```

   This will create the binary in the `out` directory.

## Usage

Run the application:
```bash
make run
```

Or build and run in one command:
```bash
make all
```

### Command Line Options

- `-v, --verbose`: Enable verbose output

## Development

### Building

```bash
make build
```

### Running Tests

```bash
make test
```

### Cleaning Build Artifacts

```bash
make clean
```
