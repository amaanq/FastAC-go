package FastAC

type AdaptiveBitModel struct {
	update_cycle, bits_until_update    uint32
	bit_0_prob, bit_0_count, bit_count uint32
}

func initAdaptiveBitModel() *AdaptiveBitModel {
	a := new(AdaptiveBitModel)
	a.reset()
	return a
}

func (a *AdaptiveBitModel) reset() {
	a.bit_0_count = 1
	a.bit_count = 2
	a.bit_0_prob = 1 << (BM__LengthShift - 1)
	a.update_cycle, a.bits_until_update = 4, 4
}

func (a *AdaptiveBitModel) Update() {
	a.bit_count += a.update_cycle
	if a.bit_count > BM__MaxCount {
		a.bit_count = (a.bit_count + 1) >> 1
		a.bit_0_count = (a.bit_0_count + 1) >> 1
		if a.bit_0_count == a.bit_count {
			a.bit_count++
		}
	}

	scale := uint32(0x80000000 / a.bit_count)
	a.bit_0_prob = (a.bit_0_count * scale) >> (31 - BM__LengthShift)

	a.update_cycle = (5 * a.update_cycle) >> 2
	if a.update_cycle > 64 {
		a.update_cycle = 64
	}
	a.bits_until_update = a.update_cycle
}
