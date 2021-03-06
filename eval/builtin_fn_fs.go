package eval

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/store/storedefs"
	"github.com/elves/elvish/util"
)

// Filesystem.

var ErrStoreNotConnected = errors.New("store not connected")

func init() {
	addToBuiltinFns([]*BuiltinFn{
		// Directory
		{"cd", cd},
		{"dir-history", dirs},

		// Path
		{"path-abs", WrapStringToStringError(filepath.Abs)},
		{"path-base", WrapStringToString(filepath.Base)},
		{"path-clean", WrapStringToString(filepath.Clean)},
		{"path-dir", WrapStringToString(filepath.Dir)},
		{"path-ext", WrapStringToString(filepath.Ext)},

		{"eval-symlinks", WrapStringToStringError(filepath.EvalSymlinks)},
		{"tilde-abbr", tildeAbbr},

		// File types
		{"-is-dir", isDir},
	})
}

func WrapStringToString(f func(string) string) BuiltinFnImpl {
	return func(ec *Frame, args []types.Value, opts map[string]types.Value) {
		TakeNoOpt(opts)
		s := mustGetOneString(args)
		ec.ports[1].Chan <- types.String(f(s))
	}
}

func WrapStringToStringError(f func(string) (string, error)) BuiltinFnImpl {
	return func(ec *Frame, args []types.Value, opts map[string]types.Value) {
		TakeNoOpt(opts)
		s := mustGetOneString(args)
		result, err := f(s)
		maybeThrow(err)
		ec.ports[1].Chan <- types.String(result)
	}
}

var errMustBeOneString = errors.New("must be one string argument")

func mustGetOneString(args []types.Value) string {
	if len(args) != 1 {
		throw(errMustBeOneString)
	}
	s, ok := args[0].(types.String)
	if !ok {
		throw(errMustBeOneString)
	}
	return string(s)
}

func cd(ec *Frame, args []types.Value, opts map[string]types.Value) {
	TakeNoOpt(opts)

	var dir string
	if len(args) == 0 {
		dir = mustGetHome("")
	} else if len(args) == 1 {
		dir = types.ToString(args[0])
	} else {
		throw(ErrArgs)
	}

	cdInner(dir, ec)
}

func cdInner(dir string, ec *Frame) {
	maybeThrow(Chdir(dir, ec.DaemonClient))
}

var dirDescriptor = types.NewStructDescriptor("path", "score")

func newDirStruct(path string, score float64) *types.Struct {
	return types.NewStruct(dirDescriptor,
		[]types.Value{types.String(path), floatToString(score)})
}

func dirs(ec *Frame, args []types.Value, opts map[string]types.Value) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	if ec.DaemonClient == nil {
		throw(ErrStoreNotConnected)
	}
	dirs, err := ec.DaemonClient.Dirs(storedefs.NoBlacklist)
	if err != nil {
		throw(errors.New("store error: " + err.Error()))
	}
	out := ec.ports[1].Chan
	for _, dir := range dirs {
		out <- newDirStruct(dir.Path, dir.Score)
	}
}

func tildeAbbr(ec *Frame, args []types.Value, opts map[string]types.Value) {
	var pathv types.String
	ScanArgs(args, &pathv)
	path := string(pathv)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	out <- types.String(util.TildeAbbr(path))
}

func isDir(ec *Frame, args []types.Value, opts map[string]types.Value) {
	var pathv types.String
	ScanArgs(args, &pathv)
	path := string(pathv)
	TakeNoOpt(opts)

	ec.OutputChan() <- types.Bool(isDirInner(path))
}

func isDirInner(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.Mode().IsDir()
}
