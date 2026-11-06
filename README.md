# Metacairn

**Write metadata to images, PDFs and office documents from the command line.**

Metacairn adds EXIF metadata to images and native document properties to PDF and
OOXML files. It always validates the result and writes through a temporary file,
so a failed operation does not leave a half-written document behind.

The name follows the geological thread of [Subsoil](https://github.com/subsll/subsoil)
and [Stratum](https://github.com/subsll/stratum): a cairn is a marker, and
Metacairn marks files with structured metadata.

## Supported formats

| Format | Extensions | Metadata written |
| --- | --- | --- |
| JPEG | `.jpg`, `.jpeg` | EXIF, IPTC and XMP fields through GoMetadata |
| PNG | `.png` | EXIF, IPTC and XMP fields through GoMetadata |
| TIFF | `.tif`, `.tiff` | EXIF, IPTC and XMP fields through GoMetadata |
| WebP | `.webp` | EXIF and XMP fields through GoMetadata |
| PDF | `.pdf` | PDF Info dictionary: Title, Author, Subject, Creator |
| Word | `.docx` | OOXML core properties |
| Excel | `.xlsx` | OOXML core properties |
| PowerPoint | `.pptx` | OOXML core properties |

PDF and OOXML do not use EXIF. Metacairn writes their native metadata formats.

## Install

Requires Go 1.27 or newer:

```sh
go install github.com/subsll/metacairn/cmd/metacairn@latest
```

If `metacairn` is not found after installation, add Go's binary directory to
your `PATH`:

```sh
# macOS/zsh
echo 'export PATH="$(go env GOPATH)/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
rehash
```

```sh
# Linux/bash
echo 'export PATH="$(go env GOPATH)/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
hash -r
```

`go install` uses `$(go env GOBIN)` when set; otherwise it uses
`$(go env GOPATH)/bin`.

## Usage

### Images

```sh
metacairn set photo.jpg \
  --output photo-tagged.jpg \
  --author "Jane Doe" \
  --copyright "CC BY 4.0" \
  --description "Amsterdam canal at dusk" \
  --keywords "amsterdam,canal,photo" \
  --gps "52.3676,4.9041" \
  --date "2026-11-05T18:30:00Z" \
  --make "Canon" \
  --model "Canon EOS 5D Mark IV" \
  --lens "EF24-70mm f/2.8L II USM" \
  --iso 100 \
  --aperture 2.8 \
  --exposure "1/250" \
  --focal-length 50
```

### PDF

```sh
metacairn set report.pdf \
  --output report-tagged.pdf \
  --title "Investigation report" \
  --author "Jane Doe" \
  --subject "Case 2026-001" \
  --software "Metacairn"
```

PDF supports **Title, Author, Subject, Description (stored as Subject), Creator
and Software (stored as Creator)**. PDF keywords and dates are not writable by
the current backend and are rejected rather than silently ignored.

### DOCX / XLSX / PPTX

```sh
metacairn set report.docx \
  --output report-tagged.docx \
  --title "Investigation report" \
  --author "Jane Doe" \
  --subject "Case 2026-001" \
  --keywords "case,report" \
  --description "Working copy" \
  --category "Research" \
  --last-modified-by "Jane Doe" \
  --created "2026-11-05T12:00:00Z"
```

### Reusable profiles

A profile is a JSON file containing metadata that can be applied to many
files. Load one or more profiles with `--profile`, or load every `.json` file
in a directory with `--profiles`:

```json
{
  "name": "my-camera",
  "description": "Metadata from my own camera",
  "fields": {
    "make": "Canon",
    "model": "Canon EOS 5D Mark IV",
    "lens": "EF24-70mm f/2.8L II USM",
    "iso": 100,
    "aperture": 2.8,
    "exposure": "1/250",
    "focal_length": "50",
    "date": "2024-06-15T14:32:00Z",
    "gps": "52.3676,4.9041",
    "orientation": 1
  }
}
```

```sh
metacairn set photo.jpg \
  --profile my-camera.json \
  --output photo-tagged.jpg
```

Profiles are merged in order, then explicit command-line fields override them,
and `--set key=value` values are applied last. A directory passed to
`--profiles` is read in sorted filename order.

The repository includes a synthetic demo profile:

```sh
metacairn set test-image.jpg \
  --profile examples/profiles/canon-eos-5d-mark-iv.demo.json \
  --output test-image-tagged.jpg
```

That demo populates Canon-style camera fields so metadata pipelines can be
tested. **It does not establish that an image was taken with a Canon camera.**
Metacairn does not remove or rewrite XMP `DigitalSourceType`, C2PA/Content
Credentials or other provenance records already in the file. Adding EXIF fields
cannot turn an edited or generated image into a camera original, and should not
be used to misrepresent provenance.

### In-place updates

`--in-place` writes a `.bak` copy next to the original before replacing it:

```sh
metacairn set photo.jpg --in-place --author "Jane Doe"
# photo.jpg is updated
# photo.jpg.bak is the untouched original
```

### Inspect

```sh
metacairn get photo-tagged.jpg
metacairn get report-tagged.pdf --json
```

### Generic fields

Every format-specific convenience flag has a `--set key=value` equivalent:

```sh
metacairn set photo.jpg --output tagged.jpg \
  --set author="Jane Doe" \
  --set keywords="one,two" \
  --set gps="52.3676,4.9041"
```

Common aliases include `artist`, `creator`, `caption`, `imagedescription`,
`rights`, `lens`, `datetime`, `location` and `coordinates`.

## Field support

| Field | Images | PDF | OOXML |
| --- | :---: | :---: | :---: |
| `title` | yes (caption) | yes | yes |
| `author` | yes | yes | yes |
| `subject` | yes (caption) | yes | yes |
| `description` | yes (caption) | yes (Subject) | yes |
| `keywords` | yes | — | yes |
| `copyright` | yes | — | — |
| `creator` | — | yes | — |
| `software` | — | yes (Creator) | — |
| `make`, `model`, `lens` | yes | — | — |
| `iso`, `aperture`, `exposure`, `focal-length` | yes | — | — |
| `date` | yes (capture date) | — | yes (created + modified) |
| `gps` | yes | — | — |
| `orientation` | yes | — | — |
| `category` | — | — | yes |
| `last-modified-by` | — | — | yes |
| `content-status` | — | — | yes |
| `created`, `modified` | — | — | yes |

## Safety and limits

Metacairn is a metadata writer, not a forensic authenticity tool.

- Output is required unless `--in-place` is used. Existing output files are
  refused unless `--force` is supplied.
- In-place changes create a `.bak` backup; an existing backup is not overwritten
  unless `--force` is supplied.
- Every write goes to a temporary file, is validated, and is then renamed into
  place. File permissions are preserved.
- PDF rewriting and ZIP rewriting change bytes outside the metadata fields. This
  may invalidate digital signatures. Make a copy and verify signed documents
  independently.
- Image writing supports common fields through GoMetadata, not arbitrary EXIF
  tag IDs or MakerNote edits.
- Profiles are metadata templates, not provenance. Camera metadata can make a
  file look camera-captured to simple field checks, but it does not prove how
  the image was made. Do not use synthetic profiles to misrepresent an image.
- OOXML support covers the standard `docProps/core.xml` properties; custom
  document properties are left untouched.
- Encrypted PDFs are not supported.
- Adding metadata does not prove that a document, image or capture is authentic.
  Metadata can be edited by anyone with the right tools.

## Build and test

```sh
go test ./...
go test -race ./...
go vet ./...
make linux
```

The test suite creates temporary JPEG, PNG, PDF and DOCX fixtures and covers
round-trips through each writer. It does not need network access.

## Dependencies

- [`github.com/FlavioCFOliveira/GoMetadata`](https://github.com/FlavioCFOliveira/GoMetadata) — MIT
- [`github.com/pdfcpu/pdfcpu`](https://github.com/pdfcpu/pdfcpu) — Apache-2.0

## License

MIT — see [LICENSE](LICENSE).
