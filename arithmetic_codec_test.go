package FastAC

import (
	"testing"
)

func TestArithmeticCodec_SetBuffer(t *testing.T) {
	type args struct {
		max_code_bytes uint32
		user_buffer    []byte
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "something", args: args{max_code_bytes: 16, user_buffer: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}}},
		{name: "nil", args: args{max_code_bytes: 0, user_buffer: nil}},
		{name: "empty", args: args{max_code_bytes: 0, user_buffer: []byte{}}}, // want panic here...
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initArithmeticCodec(tt.args.max_code_bytes, tt.args.user_buffer)
		})
	}
}

func TestArithmeticCodec_PropagateCarry(t *testing.T) {
	type args struct {
		max_code_bytes uint32
		user_buffer    []byte
	}
	tests := []struct {
		name string
		args args
	}{
		{"16 FFs to propagate", args{max_code_bytes: 16, user_buffer: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := initArithmeticCodec(tt.args.max_code_bytes, tt.args.user_buffer)
			a.ac_pointer = a.code_buffer
			a.PropagateCarry()
		})
	}
}

func BenchmarkArithmeticCodec_PropagateCarry(b *testing.B) {
	type args struct {
		max_code_bytes uint32
		user_buffer    []byte
	}
	test := args{max_code_bytes: 16, user_buffer: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}}
	// t.Run(test.name, func(t *testing.T) {
	// 	a := initArithmeticCodec(tt.args.max_code_bytes, tt.args.user_buffer)
	// 	a.ac_pointer = a.code_buffer
	// 	a.PropagateCarry2()
	// })
	for i := 0; i < b.N; i++ {
		a := initArithmeticCodec(test.max_code_bytes, test.user_buffer)
		a.ac_pointer = a.code_buffer
		a.PropagateCarry()
	}
}

func BenchmarkArithmeticCodec_PropagateCarry2(b *testing.B) {
	type args struct {
		max_code_bytes uint32
		user_buffer    []byte
	}
	test := args{max_code_bytes: 16, user_buffer: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}}

	for i := 0; i < b.N; i++ {
		a := initArithmeticCodec(test.max_code_bytes, test.user_buffer)
		a.ac_pointer = a.code_buffer
		a.deprecatedPropagateCarry2()
	}
}
