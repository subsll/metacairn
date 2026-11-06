# Changelog

All notable changes to Metacairn are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project uses
[Semantic Versioning](https://semver.org/).

## [0.1.0] — 2026-11-06

### Added

- CLI `set` and `get` commands for writing and inspecting metadata.
- JPEG, PNG, TIFF and WebP EXIF/IPTC/XMP metadata through GoMetadata.
- PDF Info dictionary support for Title, Author, Subject, Creator and Software.
- OOXML core-property support for DOCX, XLSX and PPTX.
- Atomic temporary-file writes, output validation and preserved file modes.
- In-place mode with `.bak` backups.
- Generic `--set key=value` fields and per-field convenience flags.
- Round-trip tests for images, PDF and DOCX, plus CLI tests.
