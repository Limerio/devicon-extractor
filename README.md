# Dev Icon Extractor

A Go application that extracts and organizes SVG icons from the [devicons/devicon](https://github.com/devicons/devicon) repository for my personal use. And I decided to publish it for everyone.

## Overview

This tool automates the process of:

1. Cloning the devicons repository
2. Cleaning up unnecessary files (keeping only SVG icons)
3. Extracting and organizing icons with intelligent selection priorities
4. Renaming icons according to their technology name

## Icon Selection Priority

The extractor follows this priority order when multiple SVG files exist for a technology:

1. **"original" version** - Preferred when available
2. **Exact tech name match** - Files matching the technology name
3. **"plain" version** - Simple, clean designs
4. **Single file** - If only one SVG exists, use it
5. **First available** - Fallback to the first file found

## Prerequisites

- **Go 1.21+**: Required to run the application
- **Git**: Must be installed and available in PATH for repository cloning

## Installation

1. Clone this repository:

   ```bash
   git clone https://github.com/Limerio/devicons-extractor.git
   cd devicons-extractor
   ```

2. Build the application:
   ```bash
   go build -o icon-extractor
   ```

## Usage

Run the extractor:

```bash
# Using go run
go run .

# Or using the built binary
./icon-extractor
```

The application will:

1. Clone the devicons repository to `devicon_clone/`
2. Extract SVG icons to `extracted_icons/`
3. Clean up temporary files
4. Display a summary of processed icons

## Output Structure

After extraction, you'll find all icons in the `extracted_icons/` directory:

```
extracted_icons/
├── javascript.svg
├── python.svg
├── react.svg
├── docker.svg
└── ... (more icons)
```

Each icon is named according to its technology (e.g., `javascript.svg`, `python.svg`).

## Project Structure

```
.
├── main.go           # Application entry point
├── config.go         # Configuration constants
├── extractor.go      # Core extraction logic
├── file_utils.go     # File operation utilities
├── go.mod           # Go module definition
├── README.md        # This file
├── .gitignore       # Git ignore patterns
├── devicon_clone/   # Temporary clone directory (auto-removed)
└── extracted_icons/ # Output directory with organized icons
```

# Configuration

Constants in `config.go` can be modified for customization:

```go
const (
    DeviconRepo    = "https://github.com/devicons/devicon.git"  // Source repository
    CloneDir       = "devicon_clone"                             // Temporary directory
    IconsDir       = "icons"                                     // Icons subdirectory
    OutputDir      = "extracted_icons"                           // Output directory
    DirPermissions = 0755                                        // Directory permissions
)
```

## Development

### Building

```bash
go build -o icon-extractor
```

### Code formatting

```bash
go fmt ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request
