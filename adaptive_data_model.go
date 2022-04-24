package FastAC

type AdaptiveDataModel struct {
	distribution []uint

	distribution_index, symbol_count_index, decoder_table_index uint // indexes

	total_count, update_cycle, symbols_until_update uint

	data_symbols, last_symbol, table_size, table_shift uint
}

func NewAdaptiveDataModel() *AdaptiveDataModel {
	return &AdaptiveDataModel{
		data_symbols: 0,
		distribution: make([]uint, 0),
	}
}

func NewAdaptiveDataModelWithSyms(number_of_symbols uint) {
	a := &AdaptiveDataModel{
		data_symbols: 0,
		distribution: make([]uint, 0),
	}
	a.SetAlphabet(number_of_symbols)
}

func (a *AdaptiveDataModel) Reset() {
	if a.data_symbols == 0 {
		return
	}

	a.total_count = 0
	a.update_cycle = a.data_symbols
	for k := uint(0); k < a.data_symbols; k++ {
		a.distribution[a.symbol_count_index+k] = 1
	}
	a.Update(false)
	a.symbols_until_update, a.update_cycle = (a.data_symbols+6)>>1, (a.data_symbols+6)>>1
}

func (a *AdaptiveDataModel) SetAlphabet(number_of_symbols uint) {
	if number_of_symbols < 2 || number_of_symbols > (1<<11) {
		AC_Error("invalid number of data symbols")
	}

	if a.data_symbols != number_of_symbols {
		a.data_symbols = number_of_symbols
		a.last_symbol = a.data_symbols - 1
		a.distribution = nil

		if a.data_symbols > 16 {
			table_bits := uint(3)
			for a.data_symbols > (1 << (table_bits + 2)) {
				table_bits++
			}
			a.table_size = (1 << table_bits) + 4
			a.table_shift = DM__LengthShift - table_bits
			a.distribution = make([]uint, 2*a.data_symbols+a.table_size+6)
			a.decoder_table_index = 2 * a.data_symbols
		} else {
			a.decoder_table_index = 0
			a.table_size, a.table_shift = 0, 0
			a.distribution = make([]uint, 2*a.data_symbols)
		}
		a.symbol_count_index = a.data_symbols

		a.Reset()
	}
}

func (a *AdaptiveDataModel) Update(from_encoder bool) {
	a.total_count += a.update_cycle
	if a.total_count > DM__MaxCount {
		a.total_count = 0
		for n := uint(0); n < a.data_symbols; n++ {
			a.distribution[a.symbol_count_index+n] = (a.distribution[a.symbol_count_index+n] + 1) >> 1
			a.total_count += a.distribution[a.symbol_count_index+n]
		}
	}

	k, sum, s := uint(0), uint(0), uint(0)
	scale := uint(0x80000000 / a.total_count)

	if from_encoder || a.table_size == 0 {
		for k = uint(0); k < a.data_symbols; k++ {
			a.distribution[k] = (scale * sum) >> (31 - DM__LengthShift)
			sum += a.distribution[a.symbol_count_index+k]
		}
	} else {
		for k = uint(0); k < a.data_symbols; k++ {
			a.distribution[k] = (scale * sum) >> (31 - DM__LengthShift)
			sum += a.distribution[a.symbol_count_index+k]
			w := a.distribution[k] >> a.table_shift
			for s < w {
				s++
				a.distribution[a.decoder_table_index+s] = k - 1
			}
		}
		a.distribution[a.decoder_table_index] = 0
		for s <= a.table_size {
			s++
			a.distribution[a.decoder_table_index+s] = a.data_symbols - 1
		}
	}

	a.update_cycle = (5 * a.update_cycle) >> 2
	max_cycle := uint((a.data_symbols + 6) << 3)
	if a.update_cycle > max_cycle {
		a.update_cycle = max_cycle
	}
	a.symbols_until_update = a.update_cycle
}
