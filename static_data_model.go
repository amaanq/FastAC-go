package FastAC

type StaticDataModel struct {
	distribution, decoder_table []uint32

	data_symbols, last_symbol, table_size, table_shift uint32
}

func initStaticDataModel() *StaticDataModel {
	return new(StaticDataModel)
}

func (sdm *StaticDataModel) SetDistribution(number_of_symbols uint32, probability []float64) {
	if number_of_symbols < 2 || number_of_symbols > (1<<11) {
		AC_Error("invalid number of data symbols")
	}

	if sdm.data_symbols != number_of_symbols {
		sdm.data_symbols = number_of_symbols
		sdm.last_symbol = sdm.data_symbols - 1
		sdm.distribution = nil

		if sdm.data_symbols > 16 {
			table_bits := uint32(3)
			for sdm.data_symbols > (1 << (table_bits + 2)) {
				table_bits++
			}
			sdm.table_size = (1 << table_bits)
			sdm.table_shift = DM__LengthShift - table_bits
			sdm.distribution = make([]uint32, sdm.data_symbols+sdm.table_size+2)
			sdm.decoder_table = sdm.distribution[sdm.data_symbols:]
		} else {
			sdm.decoder_table = nil
			sdm.table_size, sdm.table_shift = 0, 0
			sdm.distribution = make([]uint32, sdm.data_symbols)
		}
	}

	s := uint32(0)
	sum, p := 0.0, 1.0/float64(sdm.data_symbols)

	for k := uint32(0); k < sdm.data_symbols; k++ {
		if probability != nil {
			p = probability[k]
		}

		if p < 0.0001 || p > 0.9999 {
			AC_Error("invalid symbol probability")
		}

		sdm.distribution[k] = uint32(sum * (1 << DM__LengthShift))
		sum += p
		if sdm.table_size == 0 {
			continue
		}
		w := sdm.distribution[k] >> sdm.table_shift
		for s < w {
			s++
			sdm.decoder_table[s] = k - 1
		}
	}

	if sdm.table_size != 0 {
		sdm.decoder_table[0] = 0
		for s < sdm.table_size {
			s++
			sdm.decoder_table[s] = sdm.data_symbols - 1
		}
	}

	if sum < 0.9999 || sum > 1.0001 {
		AC_Error("invalid probabilities")
	}
}
