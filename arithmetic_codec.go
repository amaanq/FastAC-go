package FastAC

import (
	"fmt"
	"os"
)

const (
	AC__MinLength = 0x01000000
	AC__MaxLength = 0xFFFFFFFF

	BM__LengthShift = 13
	BM__MaxCount    = 1 << BM__LengthShift

	DM__LengthShift = 15
	DM__MaxCount    = 1 << DM__LengthShift
)

type ArithmeticCodec struct {
	code_buffer, new_buffer []byte
	ac_pointer_index        uint
	base, value, length     uint
	buffer_size, mode       uint
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
// - - Static functions  - - - - - - - - - - - - - - - - - - - - - - - - - - -

func AC_Error(message string) {
	panic(fmt.Errorf("\n\n -> Arithmetic coding error: %s", message))
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
// - - Coding implementations  - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) PropagateCarry() {
	var p uint
	for p = a.ac_pointer_index - 1; a.code_buffer[p] == 0xFF; p-- {
		a.code_buffer[p] = 0
	}
	a.code_buffer[p]++
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) RenormEncInterval() {
	i := uint(0)
	for a.length < AC__MinLength {
		a.code_buffer[a.ac_pointer_index+i] = byte(a.base >> 24)
		a.base <<= 8
		a.length <<= 8
		i++
	}
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) RenormDecInterval() {
	i := uint(0)
	for a.length < AC__MinLength {
		i++
		a.value = (a.value << 8) | uint(a.code_buffer[a.ac_pointer_index+i])
		a.length <<= 8
	}
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) PutBit(bit uint) {
	a.length >>= 1
	if bit > 0 {
		init_base := a.base
		a.base += a.length
		if init_base > a.base {
			a.PropagateCarry()
		}
	}
	if a.length < AC__MinLength {
		a.RenormEncInterval()
	}
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) GetBit() bool {
	a.length >>= 1
	bit := (a.value >= a.length)
	if bit {
		a.value -= a.length
	}
	if a.length < AC__MinLength {
		a.RenormDecInterval()
	}
	return bit
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) PutBits(data, bits uint) {
	init_base := a.base
	a.length >>= bits
	a.base += data * a.length

	if init_base > a.base {
		a.PropagateCarry()
	}
	if a.length < AC__MinLength {
		a.RenormEncInterval()
	}
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) GetBits(bits uint) uint {
	a.length >>= bits
	s := a.value / a.length
	a.value -= a.length * s
	if a.length < AC__MinLength {
		a.RenormDecInterval()
	}
	return s
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) EncodeFromStaticBitModel(bit uint, M *StaticBitModel) {
	x := M.bit0prob * (a.length >> BM__LengthShift)
	if bit == 0 {
		a.length = x
	} else {
		init_base := a.base
		a.base += x
		a.length -= x
		if init_base > a.base {
			a.PropagateCarry()
		}
	}

	if a.length < AC__MinLength {
		a.RenormEncInterval()
	}
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) DecodeFromStaticBitModel(M *StaticBitModel) bool {
	x := M.bit0prob * (a.length >> BM__LengthShift)
	bit := a.value >= x

	if !bit {
		a.length = x
	} else {
		a.value -= x
		a.length -= x
	}

	if a.length < AC__MinLength {
		a.RenormDecInterval()
	}
	return bit
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) EncodeFromAdaptiveBitModel(bit uint, M *AdaptiveBitModel) {
	x := M.bit_0_prob * (a.length >> BM__LengthShift)

	if bit == 0 {
		a.length = x
		M.bit_0_count++
	} else {
		init_base := a.base
		a.base += x
		a.length -= x
		if init_base > a.base {
			a.PropagateCarry()
		}
	}

	if a.length < AC__MinLength {
		a.RenormEncInterval()
	}

	M.bits_until_update--
	if M.bits_until_update == 0 {
		M.Update()
	}
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) DecodeFromAdaptiveBitModel(M *AdaptiveBitModel) bool {
	x := M.bit_0_prob * (a.length >> BM__LengthShift)
	bit := a.value >= x

	if !bit {
		a.length = x
		M.bit_0_count++
	} else {
		a.value -= x
		a.length -= x
	}

	if a.length < AC__MinLength {
		a.RenormDecInterval()
	}

	M.bits_until_update--
	if M.bits_until_update == 0 {
		M.Update()
	}
	return bit
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) EncodeFromStaticDataModel(data uint, M *StaticDataModel) {
	x, init_base := a.base, a.base

	if data == M.last_symbol {
		x = M.distribution[data] * (a.length >> DM__LengthShift)
		a.base += x
		a.length -= x
	} else {
		a.length >>= DM__LengthShift
		x = M.distribution[data] * (a.length)
		a.base += x
		a.length = M.distribution[data+1]*a.length - x
	}

	if init_base > a.base {
		a.PropagateCarry()
	}

	if a.length < AC__MinLength {
		a.RenormEncInterval()
	}
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) DecodeFromStaticDataModel(M *StaticDataModel) uint {
	n, s, x, y := a.length, a.length, a.length, a.length

	if M.distribution != nil {
		a.length >>= DM__LengthShift
		dv := a.value / a.length
		t := dv >> M.table_shift

		s = M.distribution[M.decoder_table_index+t]
		n = M.distribution[M.decoder_table_index+t+1] + 1

		for n > s+1 {
			m := (s + n) >> 1
			if M.distribution[m] > dv {
				n = m
			} else {
				s = m
			}
		}
		x = M.distribution[s] * a.length
		if s != M.last_symbol {
			y = M.distribution[s+1] * a.length
		}
	} else {
		x, s = 0, 0
		a.length >>= DM__LengthShift
		n = M.data_symbols
		m := n >> 1

		m = (s + n) >> 1
		for m != s {
			z := a.length * M.distribution[m]
			if z > a.value {
				n = m
				y = z
			} else {
				s = m
				x = z
			}
			m = (s + n) >> 1
		}
	}

	a.value -= x
	a.length = y - x

	if a.length < AC__MinLength {
		a.RenormDecInterval()
	}

	return s
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) EncodeFromAdaptiveDataModel(data uint, M *AdaptiveDataModel) {
	x, init_base := a.base, a.base

	if data == M.last_symbol {
		x = M.distribution[data] * (a.length >> DM__LengthShift)
		a.base += x
		a.length -= x
	} else {
		x = M.distribution[data] * (a.length >> DM__LengthShift)
		a.base += x
		a.length = M.distribution[data+1]*a.length - x
	}

	if init_base > a.base {
		a.PropagateCarry()
	}

	if a.length < AC__MinLength {
		a.RenormEncInterval()
	}

	M.distribution[M.symbol_count_index+data]++
	M.symbols_until_update--
	if M.symbols_until_update == 0 {
		M.Update(true)
	}
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) DecodeFromAdaptiveDataModel(M *AdaptiveDataModel) uint {
	n, s, x, y := a.length, a.length, a.length, a.length

	if M.distribution != nil {
		a.length >>= DM__LengthShift
		dv := a.value / a.length
		t := dv >> M.table_shift

		s = M.distribution[M.decoder_table_index+t]
		n = M.distribution[M.decoder_table_index+t+1] + 1

		for n > s+1 {
			m := (s + n) >> 1
			if M.distribution[m] > dv {
				n = m
			} else {
				s = m
			}
		}

		x = M.distribution[s] * a.length
		if s != M.last_symbol {
			y = M.distribution[s+1] * a.length
		}
	} else {
		x, s = 0, 0
		a.length >>= DM__LengthShift
		n = M.data_symbols
		m := n >> 1

		m = (s + n) >> 1
		for m != s {
			z := a.length * M.distribution[m]
			if z > a.value {
				n = m
				y = z
			} else {
				s = m
				x = z
			}
			m = (s + n) >> 1
		}
	}

	a.value -= x
	a.length = y - x

	if a.length < AC__MinLength {
		a.RenormDecInterval()
	}

	M.distribution[M.symbol_count_index+s]++
	M.symbols_until_update--
	if M.symbols_until_update == 0 {
		M.Update(false)
	}
	return s
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
// - - Other Arithmetic_Codec implementations  - - - - - - - - - - - - - - - -

func NewArithmeticCodec() *ArithmeticCodec {
	return &ArithmeticCodec{
		mode:        0,
		buffer_size: 0,
		new_buffer:  make([]byte, 0),
		code_buffer: make([]byte, 0),
	}
}

func NewArithmeticCodecFromBuffer(max_code_bytes uint, user_buffer []byte) {
	a := NewArithmeticCodec()
	a.SetBuffer(max_code_bytes, user_buffer)
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) SetBuffer(max_code_bytes uint, user_buffer []byte) {
	if max_code_bytes < 16 || max_code_bytes > 0x1000000 {
		AC_Error("invalid codec buffer size")
	}
	if a.mode != 0 {
		AC_Error("cannot set buffer while encoding or decoding")
	}

	if user_buffer != nil {
		a.buffer_size = max_code_bytes
		a.code_buffer = user_buffer
		a.new_buffer = make([]byte, 0)
		return
	}

	if max_code_bytes <= a.buffer_size {
		return
	}

	a.buffer_size = max_code_bytes
	a.new_buffer = make([]byte, a.buffer_size+16)
	a.code_buffer = a.new_buffer
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) StartEncoder() {
	if a.mode != 0 {
		AC_Error("cannot start encoder")
	}
	if a.buffer_size == 0 {
		AC_Error("no code buffer set")
	}

	a.mode = uint(Encoder)
	a.base = 0
	a.length = AC__MaxLength
	a.ac_pointer_index = 0 // start of code_buffer
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) StartDecoder() {
	if a.mode != 0 {
		AC_Error("cannot start decoder")
	}
	if a.buffer_size == 0 {
		AC_Error("no code buffer set")
	}
	a.mode = uint(Decoder)
	a.length = AC__MaxLength
	a.ac_pointer_index = 3 // code_buffer + 3
	a.value = uint(a.code_buffer[0])<<24 | uint(a.code_buffer[1])<<16 | uint(a.code_buffer[2])<<8 | uint(a.code_buffer[3])
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) ReadFromFile(file *os.File) {
	// do later
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) StopEncoder() uint {
	if a.mode != uint(Encoder) {
		AC_Error("invalid to stop encoder")
	}
	a.mode = 0

	init_base := a.base

	if a.length > 2*AC__MinLength {
		a.base += AC__MinLength
		a.length = AC__MinLength >> 1
	} else {
		a.base += AC__MinLength >> 1
		a.length = AC__MinLength >> 9
	}

	if init_base > a.base {
		a.PropagateCarry()
	}

	a.RenormEncInterval()

	code_bytes := uint(a.ac_pointer_index)

	if code_bytes > a.buffer_size {
		AC_Error("code buffer overflow")
	}

	return code_bytes
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) WriteToFile(file *os.File) {
	// do later
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) StopDecoder() {
	if a.mode != uint(Decoder) {
		AC_Error("invalid to stop decoder")
	}
	a.mode = 0
}
