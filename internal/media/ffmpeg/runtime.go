// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package ffmpeg

import (
	"context"
	"io"
	"os"

	ffmpeglib "codeberg.org/gruf/go-ffmpreg/embed/ffmpeg"
	ffprobelib "codeberg.org/gruf/go-ffmpreg/embed/ffprobe"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// Use all core features required by ffmpeg / ffprobe
// (these should be the same but we OR just in case).
const corefeatures = ffprobelib.CoreFeatures |
	ffmpeglib.CoreFeatures

// Args encapsulates the passing of common
// configuration options to run an instance
// of a compiled WebAssembly module that is
// run in a typical CLI manner.
type Args struct {

	// Optional further module configuration function.
	// (e.g. to mount filesystem dir, set env vars, etc).
	Config func(wazero.ModuleConfig) wazero.ModuleConfig

	// Standard FDs.
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	// CLI args.
	Args []string
}

// shared WASM runtime instance.
var runtime wazero.Runtime

// initRuntime initializes the global wazero.Runtime,
// if already initialized this function is a no-op.
func initRuntime(ctx context.Context) error {
	if runtime != nil {
		return nil
	}

	var cache wazero.CompilationCache

	if dir := os.Getenv("GTS_WAZERO_COMPILATION_CACHE"); dir != "" {
		var err error

		// Use on-filesystem compilation cache given by env.
		cache, err = wazero.NewCompilationCacheWithDir(dir)
		if err != nil {
			return err
		}
	}

	// Prepare config with cache.
	cfg := wazero.NewRuntimeConfig()
	cfg = cfg.WithCoreFeatures(corefeatures)
	cfg = cfg.WithCompilationCache(cache)

	// Instantiate runtime with prepared config.
	rt := wazero.NewRuntimeWithConfig(ctx, cfg)

	// Instantiate wasi snapshot preview features into runtime.
	_, err := wasi_snapshot_preview1.Instantiate(ctx, rt)
	if err != nil {
		return err
	}

	// Set runtime.
	runtime = rt
	return nil
}
