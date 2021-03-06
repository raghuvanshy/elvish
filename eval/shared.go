package eval

import (
	"github.com/elves/elvish/daemon"
	"github.com/elves/elvish/eval/types"
)

type sharedVariable struct {
	store *daemon.Client
	name  string
}

func (sv sharedVariable) Set(val types.Value) error {
	return sv.store.SetSharedVar(sv.name, types.ToString(val))
}

func (sv sharedVariable) Get() types.Value {
	value, err := sv.store.SharedVar(sv.name)
	maybeThrow(err)
	return types.String(value)
}
