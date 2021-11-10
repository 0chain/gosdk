package jsbridge

import (
	"reflect"
	"syscall/js"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBinder(t *testing.T) {

	tests := []struct {
		Name string
		In   []js.Value
		Out  interface{}
		Func func() reflect.Value
	}{
		{Name: "string", Func: func() reflect.Value {
			fn := func(i string) string {
				return i
			}

			return reflect.ValueOf(fn)
		}, In: []js.Value{js.ValueOf("1")}, Out: "1"},

		{Name: "int", Func: func() reflect.Value {
			fn := func(i int) int {
				return i
			}

			return reflect.ValueOf(fn)
		}, In: []js.Value{js.ValueOf(1)}, Out: 1},
		{Name: "int32", Func: func() reflect.Value {
			fn := func(i int32) int32 {
				return i
			}

			return reflect.ValueOf(fn)
		}, In: []js.Value{js.ValueOf(1)}, Out: int32(1)},
		{Name: "int64", Func: func() reflect.Value {
			fn := func(i int64) int64 {
				return i
			}

			return reflect.ValueOf(fn)
		}, In: []js.Value{js.ValueOf(1)}, Out: int64(1)},

		{Name: "float32", Func: func() reflect.Value {
			fn := func(i float32) float32 {
				return i
			}

			return reflect.ValueOf(fn)
		}, In: []js.Value{js.ValueOf(1)}, Out: float32(1)},
		{Name: "float64", Func: func() reflect.Value {
			fn := func(i float64) float64 {
				return i
			}

			return reflect.ValueOf(fn)
		}, In: []js.Value{js.ValueOf(1)}, Out: float64(1)},
	}

	for _, it := range tests {
		t.Run(it.Name, func(test *testing.T) {
			fn := it.Func()
			b, err := NewInputBuilder(fn.Type()).Build()

			require.NoError(test, err)

			in, err := b(it.In)
			require.NoError(test, err)

			out := fn.Call(in)

			require.Equal(test, it.Out, out[0].Interface())

		})
	}

}
