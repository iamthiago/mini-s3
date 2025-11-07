# mini-s3

A (Work In Progress) lightweight, S3-like object storage system with a CLI interface, written in Go.

> **Note:** This is a personal learning project that I'm working on in my free time.
> I chose a mini-s3, as it presents several challenges that I'm interested in exploring.

## Overview

mini-s3 is a simple implementation of object storage that mimics Amazon S3's bucket/object paradigm. It provides local
filesystem-based storage with checksums for data integrity verification.

## Features and a possible Roadmap

- [x] Local filesystem
- [ ] CLI
- [ ] Metadata
- [ ] Replication
- [ ] Consensus

## Installation

### From Source

```bash
git clone https://github.com/iamthiago/mini-s3.git
cd mini-s3
make build
```

The binary will be available at `./mini-s3`

### Requirements

- Go 1.25 or higher

## Usage

### Basic Commands

```bash
# List objects in a bucket
mini-s3 list <bucket-name>

# Put an object into a bucket
mini-s3 put <bucket-name> <object-key> <file-path>

# Get an object from a bucket
mini-s3 get <bucket-name> <object-key> [output-path]

# Delete an object from a bucket
mini-s3 delete <bucket-name> <object-key>
```

### Examples

```bash
# Store a file
mini-s3 put my-bucket documents/report.pdf ./report.pdf

# Retrieve a file
mini-s3 get my-bucket documents/report.pdf ./downloaded-report.pdf

# List all objects in a bucket
mini-s3 list my-bucket

# Delete a file
mini-s3 delete my-bucket documents/report.pdf
```

## Architecture

### Project Structure

```
mini-s3/
├── cmd/                   # CLI commands and command tests
├── internal/
│   └── storage/           # Core storage implementation, checksums, and tests
├── data/                  # Default data directory for local storage
├── main.go                # Application entry point
├── Makefile               # Build and development tasks
└── go.mod                 # Go module dependencies
```
