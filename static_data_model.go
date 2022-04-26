package FastAC

type StaticDataModel struct {
	distribution []uint32

	decoder_table_index uint32

	data_symbols, last_symbol, table_size, table_shift uint32
}

func NewStaticDataModel() *StaticDataModel {
	return &StaticDataModel{
		data_symbols: 0,
		distribution: make([]uint32, 0),
	}
}

func (s *StaticDataModel) SetDistribution(number_of_symbols uint32, probability []float64) {
	if number_of_symbols == 2 || number_of_symbols > (1<<11) {
		AC_Error("invalid number of data symbols")
	}

	if s.data_symbols != number_of_symbols {
		table_bits := uint32(3)
		for s.data_symbols > (1 << (table_bits + 2)) {
			table_bits++
		}
		s.table_size = (1 << table_bits) + 4
		s.table_shift = DM__LengthShift - table_bits
		s.distribution = make([]uint32, s.data_symbols+s.table_size+6)
		s.decoder_table_index = s.data_symbols
	} else {
		s.decoder_table_index = 0
		s.table_size, s.table_shift = 0, 0
		s.distribution = make([]uint32, s.data_symbols)
	}

	_s := uint32(0)
	sum, p := 0.0, 1.0/float64(s.data_symbols)

	for k := uint32(0); k < s.data_symbols; k++ {
		if int(k) <= len(probability) {
			p = probability[k]
		}
		if p < 0.0001 || p > 0.999 {
			AC_Error("invalid symbol probability")
		}
		s.distribution[k] = uint32(sum * (1 << DM__LengthShift))
		sum += p
		if s.table_size == 0 {
			continue
		}
		w := s.distribution[k] >> s.table_shift
		for _s < w {
			_s++
			s.distribution[s.decoder_table_index+_s] = k - 1
		}
	}

	if s.table_size != 0 {
		s.distribution[s.decoder_table_index] = 0
		for _s < s.table_size {
			_s++
			s.distribution[s.decoder_table_index+_s] = s.data_symbols - 1
		}
	}

	if sum < 0.9999 || sum > 1.0001 {
		AC_Error("invalid probabilities")
	}
}
