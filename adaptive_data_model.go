package FastAC

type AdaptiveDataModel struct {
	distribution, symbol_count, decoder_table []uint32

	total_count, update_cycle, symbols_until_update uint32

	data_symbols, last_symbol, table_size, table_shift uint32
}

func initAdaptiveDataModel(number_of_symbols uint32) *AdaptiveDataModel {
	model := new(AdaptiveDataModel)
	model.SetAlphabet(number_of_symbols)
	return model
}

func (a *AdaptiveDataModel) SetAlphabet(number_of_symbols uint32) {
	if number_of_symbols < 2 || number_of_symbols > (1<<11) {
		AC_Error("invalid number of data symbols")
	}

	if a.data_symbols != number_of_symbols {
		a.data_symbols = number_of_symbols
		a.last_symbol = a.data_symbols - 1
		a.distribution = nil

		if a.data_symbols > 16 {
			table_bits := uint32(3)
			for a.data_symbols > (1 << (table_bits + 2)) {
				table_bits++
			}
			a.table_size = 1 << table_bits
			a.table_shift = DM__LengthShift - table_bits
			a.distribution = make([]uint32, 2*a.data_symbols+a.table_size+2)
			a.decoder_table = a.distribution[2*a.data_symbols:]
		} else {
			a.decoder_table = nil
			a.table_size, a.table_shift = 0, 0
			a.distribution = make([]uint32, 2*a.data_symbols)
		}
		a.symbol_count = a.distribution[a.data_symbols:]
	}
	a.Reset()
}

func (a *AdaptiveDataModel) Update(from_encoder bool) {
	a.total_count += a.update_cycle
	if a.total_count > DM__MaxCount {
		a.total_count = 0
		for n := uint32(0); n < a.data_symbols; n++ {
			a.symbol_count[n] = (a.symbol_count[n] + 1) >> 1
			a.total_count += a.symbol_count[n]
		}
	}

	var k, sum, s uint32
	scale := uint32(0x80000000 / a.total_count)

	if from_encoder || a.table_size == 0 {
		for k = uint32(0); k < a.data_symbols; k++ {
			a.distribution[k] = (scale * sum) >> (31 - DM__LengthShift)
			sum += a.symbol_count[k]
		}
	} else {
		for k = uint32(0); k < a.data_symbols; k++ {
			a.distribution[k] = (scale * sum) >> (31 - DM__LengthShift)
			sum += a.symbol_count[k]
			w := a.distribution[k] >> a.table_shift
			for s < w {
				s++
				a.decoder_table[s] = k - 1
			}
		}
		a.decoder_table[0] = 0
		for s <= a.table_size {
			s++
			a.decoder_table[s] = a.data_symbols - 1
		}
	}

	a.update_cycle = (5 * a.update_cycle) >> 2
	max_cycle := uint32((a.data_symbols + 6) << 3)
	if a.update_cycle > max_cycle {
		a.update_cycle = max_cycle
	}
	a.symbols_until_update = a.update_cycle
}

func (a *AdaptiveDataModel) Reset() {
	if a.data_symbols == 0 {
		return
	}

	a.total_count = 0
	a.update_cycle = a.data_symbols
	for k := uint32(0); k < a.data_symbols; k++ {
		a.symbol_count[k] = 1
	}
	a.Update(false)
	a.symbols_until_update, a.update_cycle = (a.data_symbols+6)>>1, (a.data_symbols+6)>>1
}
