package exercises

import "embed"

// FS contains exercise metadata, instructions, starter templates, and tests.
// Reference solutions intentionally are not embedded in the CLI binary.
//
//go:embed */exercise.json */README.md */starter/* */tests/*
var FS embed.FS
