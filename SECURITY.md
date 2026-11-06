# Security policy

## Reporting

Please report security issues privately to the maintainer rather than opening a
public issue. Include the affected version, platform and a minimal synthetic
reproducer. Do not send sensitive case material.

## Scope

Metacairn rewrites metadata in local files. It does not fetch remote content,
execute macros, decrypt documents or validate authenticity. PDF and OOXML
rewrites may invalidate digital signatures; treat signed files accordingly.
