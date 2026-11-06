# Changelog

All notable changes to Metacairn are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project uses
[Semantic Versioning](https://semver.org/).

## [0.2.0] — 2026-11-06

### Added

- Reusable JSON metadata profiles with `--profile` and `--profiles`.
- Profile directories merge all `.json` profiles in sorted order.
- Camera fields: ISO, aperture, exposure time and focal length.
- Canon EOS 5D Mark IV demo profile under `examples/profiles/`.

### Changed

- Command-line fields override profile values; `--set` values are applied last.
- Documentation now warns that profiles do not create provenance and do not
  remove XMP/C2PA metadata.

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
