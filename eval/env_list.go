package eval

import (
	"errors"
	"os"
	"strings"
	"sync"

	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/eval/vartypes"
	"github.com/xiaq/persistent/vector"
)

var (
	pathListSeparator = string(os.PathListSeparator)
	forbiddenInPath   = pathListSeparator + "\x00"
)

// Errors
var (
	ErrCanOnlyAssignList          = errors.New("can only assign compatible values")
	ErrPathMustBeString           = errors.New("path must be string")
	ErrPathCannotContainColonZero = errors.New(`path cannot contain colon or \0`)
)

// EnvList is a variable whose value is constructed from an environment variable
// by splitting at pathListSeparator. Changes to it are also propagated to the
// corresponding environment variable. Its elements cannot contain
// pathListSeparator or \0; attempting to put any in its elements will result in
// an error.
type EnvList struct {
	sync.RWMutex
	envName    string
	cacheFor   string
	cacheValue types.Value
}

var (
	_ vartypes.Variable = (*EnvList)(nil)
)

// Get returns a Value for an EnvPathList.
func (envli *EnvList) Get() types.Value {
	envli.Lock()
	defer envli.Unlock()

	value := os.Getenv(envli.envName)
	if value == envli.cacheFor {
		return envli.cacheValue
	}
	envli.cacheFor = value
	v := vector.Empty
	for _, path := range strings.Split(value, pathListSeparator) {
		v = v.Cons(types.String(path))
	}
	envli.cacheValue = types.NewList(v)
	return envli.cacheValue
}

// Set sets an EnvPathList. The underlying environment variable is set.
func (envli *EnvList) Set(v types.Value) error {
	iterator, ok := v.(types.Iterator)
	if !ok {
		return ErrCanOnlyAssignList
	}
	var paths []string
	var err error
	iterator.Iterate(func(v types.Value) bool {
		s, ok := v.(types.String)
		if !ok {
			err = ErrPathMustBeString
			return false
		}
		path := string(s)
		if strings.ContainsAny(path, forbiddenInPath) {
			err = ErrPathCannotContainColonZero
			return false
		}
		paths = append(paths, string(s))
		return true
	})

	if err != nil {
		return err
	}

	envli.Lock()
	defer envli.Unlock()
	os.Setenv(envli.envName, strings.Join(paths, pathListSeparator))
	return nil
}
