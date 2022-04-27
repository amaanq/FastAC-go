package FastAC

import (
	"fmt"
	"os"
)

const (
	AC__MinLength = 0x01000000
	AC__MaxLength = 0xFFFFFFFF

	BM__LengthShift = 13                   // Bit Model Length Shift
	BM__MaxCount    = 1 << BM__LengthShift // 1 shifted left BM__LengthShift

	DM__LengthShift = 15                   // Data Model Length Shift
	DM__MaxCount    = 1 << DM__LengthShift // 1 shifted left DM__LengthShift
)

type ArithmeticCodec struct {
	// The ac pointer is really there to just count backwards... so we can just use code_buffer
	code_buffer, new_buffer, ac_pointer []byte
	base, value, length                 uint32
	buffer_size                         uint32
	mode                                Mode
}

func AC_Error(message string) {
	panic(fmt.Errorf("\n\n -> Arithmetic coding error: %s", message))
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
// - - Static functions  - - - - - - - - - - - - - - - - - - - - - - - - - - -

func initArithmeticCodec(max_code_bytes uint32, user_buffer []byte) *ArithmeticCodec {
	codec := new(ArithmeticCodec)
	codec.SetBuffer(max_code_bytes, user_buffer)
	return codec
}

func (a *ArithmeticCodec) SetBuffer(max_code_bytes uint32, user_buffer []byte) {
	if (max_code_bytes < 16 || max_code_bytes > 0x1000000) && user_buffer != nil {
		AC_Error("invalid codec buffer size: " + fmt.Sprint(max_code_bytes))
	}
	if a.mode != Undefined {
		AC_Error("cannot set buffer while encoding or decoding")
	}

	if user_buffer != nil {
		a.buffer_size = max_code_bytes
		a.code_buffer = user_buffer
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
// - - Coding implementations  - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) PropagateCarry() {
	var p uint32
	for p := len(a.ac_pointer) - 1; a.ac_pointer[p] == 0xFF && p != 0; p-- {
		a.ac_pointer[p] = 0
	}
	a.ac_pointer[p]++
}

// Written for experimental purposes to see if this implementation is faster and uses less memory (side note, it's about the same)
func (a *ArithmeticCodec) deprecatedPropagateCarry2() {
	var p []byte
	for p = a.ac_pointer; p[len(p)-1] == 0xFF && len(p) != 1; p = p[:len(p)-1] {
		p[len(p)-1] = 0
	}
	p[len(p)-1]++
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) RenormEncInterval() {
	i := uint32(0)
	for cont := true; cont; cont = a.length < AC__MinLength { // eval at least once
		a.ac_pointer[i] = byte(a.base >> 24)
		a.base <<= 8
		a.length <<= 8
		i++
	}
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) RenormDecInterval() {
	i := uint32(0)
	for cont := true; cont; cont = a.length < AC__MinLength {
		i++
		a.value = (a.value << 8) | uint32(a.ac_pointer[i])
		a.length <<= 8
	}
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) PutBit(bit uint32) {
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

func (a *ArithmeticCodec) PutBits(data, bits uint32) {
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

func (a *ArithmeticCodec) GetBits(bits uint32) uint32 {
	a.length >>= bits
	s := a.value / a.length
	a.value -= a.length * s
	if a.length < AC__MinLength {
		a.RenormDecInterval()
	}
	return s
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) Encode_StaticBitModel(bit uint32, M *StaticBitModel) {
	x := M.bit_0_prob * (a.length >> BM__LengthShift)
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

func (a *ArithmeticCodec) Decode_StaticBitModel(M *StaticBitModel) uint32 {
	x := M.bit_0_prob * (a.length >> BM__LengthShift)
	bit := uint32(0)
	if a.value >= x {
		bit = 1
	}

	if bit == 0 {
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

func (a *ArithmeticCodec) Encode_AdaptiveBitModel(bit uint32, M *AdaptiveBitModel) {
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

func (a *ArithmeticCodec) Decode_AdaptiveBitModel(M *AdaptiveBitModel) uint32 {
	x := M.bit_0_prob * (a.length >> BM__LengthShift)
	bit := uint32(0)
	if a.value >= x {
		bit = 1
	}

	if bit == 0 {
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

func (a *ArithmeticCodec) Encode_StaticDataModel(data uint32, M *StaticDataModel) {
	var x uint32
	var init_base uint32 = a.base

	if data == M.last_symbol {
		x = M.distribution[data] * (a.length >> DM__LengthShift)
		a.base += x
		a.length -= x
	} else {
		a.length >>= DM__LengthShift
		x = M.distribution[data] * a.length
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

func (a *ArithmeticCodec) Decode_StaticDataModel(M *StaticDataModel) uint32 {
	var n, s, x uint32
	var y uint32 = a.length

	if M.decoder_table != nil {
		a.length >>= DM__LengthShift
		dv := a.value / a.length
		t := dv >> M.table_shift

		s = M.decoder_table[t]
		n = M.decoder_table[t+1] + 1

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

		for cont := true; cont; cont = m != s {
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

func (a *ArithmeticCodec) Encode_AdaptiveDataModel(data uint32, M *AdaptiveDataModel) {
	var x uint32
	var init_base uint32 = a.base

	if data == M.last_symbol {
		x = M.distribution[data] * (a.length >> DM__LengthShift)
		a.base += x
		a.length -= x
	} else {
		a.length >>= DM__LengthShift
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

	M.symbol_count[data]++
	M.symbols_until_update--
	if M.symbols_until_update == 0 {
		M.Update(true)
	}
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) Decode_AdaptiveDataModel(M *AdaptiveDataModel) uint32 {
	var n, s, x uint32
	var y uint32 = a.length

	if M.decoder_table != nil {
		a.length >>= DM__LengthShift
		dv := a.value / a.length
		t := dv >> M.table_shift

		s = M.decoder_table[t]
		n = M.decoder_table[t+1] + 1

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

		for cont := true; cont; cont = m != s {
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

	M.symbol_count[s]++
	M.symbols_until_update--
	if M.symbols_until_update == 0 {
		M.Update(false)
	}
	return s
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
// - - Other Arithmetic_Codec implementations  - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) StartEncoder() {
	if a.mode != Undefined {
		AC_Error("cannot start encoder")
	}
	if a.buffer_size == 0 {
		AC_Error("no code buffer set")
	}

	a.mode = Encoder
	a.base = 0
	a.length = AC__MaxLength
	a.ac_pointer = a.code_buffer
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) StartDecoder() {
	if a.mode != Undefined {
		AC_Error("cannot start decoder")
	}
	if a.buffer_size == 0 {
		AC_Error("no code buffer set")
	}
	a.mode = Decoder
	a.length = AC__MaxLength
	a.ac_pointer = a.code_buffer[3:] // code_buffer + 3
	a.value = uint32(a.code_buffer[0])<<24 | uint32(a.code_buffer[1])<<16 | uint32(a.code_buffer[2])<<8 | uint32(a.code_buffer[3])
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) ReadFromFile(file *os.File) {
	var shift, code_bytes uint32 = 0, 0
	var file_byte int32

	for cont := true; cont; cont = file_byte&0x80 != 0 {
		singleByte := make([]byte, 1)
		n, err := file.Read(singleByte)
		if err != nil {
			AC_Error("error reading from file: " + err.Error())
		}
		if n != 1 {
			AC_Error("error reading from file: unexpected EOF | n = " + fmt.Sprint(n))
		}
		file_byte = int32(singleByte[0])
		code_bytes |= uint32(file_byte&0x7F) << shift
		shift += 7
	}
	if code_bytes > a.buffer_size {
		AC_Error("code buffer overflow")
	}

	a.StartDecoder()
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) StopEncoder() uint32 {
	if a.mode != Encoder {
		AC_Error("invalid to stop encoder")
	}
	a.mode = Undefined

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

	code_bytes := uint32(len(a.code_buffer) - len(a.ac_pointer))

	if code_bytes > a.buffer_size {
		AC_Error("code buffer overflow")
	}

	return code_bytes
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) WriteToFile(file *os.File) uint32 {
	var header_bytes, code_bytes uint32 = 0, a.StopEncoder()
	var nb = code_bytes

	for cont := true; cont; cont = nb > 0 {
		file_byte := int(nb & 0x7F)
		nb >>= 7
		if nb > 0 {
			file_byte |= 0x80
		}
		n, err := file.Write([]byte{byte(file_byte)})
		if err != nil {
			AC_Error("cannot write compressed data to file: " + err.Error())
		}
		if n != 1 {
			AC_Error("cannot write compressed data to file: unexpected EOF | n = " + fmt.Sprint(n))
		}
		header_bytes++
	}
	n, err := file.Write(a.code_buffer[:code_bytes])
	if err != nil {
		AC_Error("cannot write compressed data to file: " + err.Error())
	}
	if n != int(code_bytes) {
		AC_Error("cannot write compressed data to file: unexpected EOF | n = " + fmt.Sprint(n))
	}

	return header_bytes + code_bytes
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (a *ArithmeticCodec) StopDecoder() {
	if a.mode != Decoder {
		AC_Error("invalid to stop decoder")
	}
	a.mode = Undefined
}
