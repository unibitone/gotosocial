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
	"sync"

	ffmpeglib "codeberg.org/gruf/go-ffmpreg/embed/ffmpeg"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/sys"
)

var (
	// ffmpegRunner limits the number of
	// ffmpeg WebAssembly instances that
	// may be concurrently running, in
	// order to reduce memory usage.
	ffmpegRunner runner

	// ffmpeg compiled WASM.
	ffmpeg wazero.CompiledModule

	// Number of times ffmpeg
	// compiled WASM has run.
	ffmpegRunCount int

	// Sync for updating run count
	// and recompiling ffmpeg.
	ffmpegM sync.Mutex
)

// InitFfmpeg precompiles the ffmpeg WebAssembly source into memory and
// prepares the runner to only allow max given concurrent running instances.
func InitFfmpeg(ctx context.Context, max int) error {

	// Ensure runner initialized.
	ffmpegRunner.Init(max)

	// Ensure runtime initialized.
	if err := initRuntime(ctx); err != nil {
		return err
	}

	// Ensure ffmpeg compiled.
	if ffmpeg == nil {
		return compileFfmpeg(ctx)
	}

	return nil
}

// compileFfmpeg ensures the ffmpeg WebAssembly
// module has been pre-compiled into memory.
func compileFfmpeg(ctx context.Context) error {
	var err error
	ffmpeg, err = runtime.CompileModule(ctx, ffmpeglib.B)
	return err
}

// Ffmpeg runs the given arguments with an instance of ffmpeg.
func Ffmpeg(ctx context.Context, args Args) (uint32, error) {
	return ffmpegRunner.Run(ctx, func() (uint32, error) {

		// Update run count + check if we
		// need to recompile the module.
		ffmpegM.Lock()
		{
			ffmpegRunCount++
			if ffmpegRunCount > 500 {
				// Over our threshold of runs, close
				// current compiled module and recompile.
				if err := ffmpeg.Close(ctx); err != nil {
					ffmpegM.Unlock()
					return 0, err
				}

				if err := compileFfmpeg(ctx); err != nil {
					ffmpegM.Unlock()
					return 0, err
				}

				ffmpegRunCount = 0
			}
		}
		ffmpegM.Unlock()

		// Prefix module name as argv0 to args.
		cargs := make([]string, len(args.Args)+1)
		copy(cargs[1:], args.Args)
		cargs[0] = "ffmpeg"

		// Create base module config.
		modcfg := wazero.NewModuleConfig()
		modcfg = modcfg.WithArgs(cargs...)
		modcfg = modcfg.WithStdin(args.Stdin)
		modcfg = modcfg.WithStdout(args.Stdout)
		modcfg = modcfg.WithStderr(args.Stderr)

		if args.Config != nil {
			// Pass through config fn.
			modcfg = args.Config(modcfg)
		}

		// Instantiate the module from precompiled wasm module data.
		mod, err := runtime.InstantiateModule(ctx, ffmpeg, modcfg)

		if mod != nil {
			// Ensure closed.
			if err := mod.Close(ctx); err != nil {
				log.Errorf(ctx, "error closing: %v", err)
			}
		}

		// Try extract exit code.
		switch err := err.(type) {
		case *sys.ExitError:
			return err.ExitCode(), nil
		default:
			return 0, err
		}
	})
}
