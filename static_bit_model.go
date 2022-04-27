package FastAC

type StaticBitModel struct {
	bit_0_prob uint32
}

func initStaticBitModel() *StaticBitModel {
	return &StaticBitModel{
		bit_0_prob: 1 << (BM__LengthShift - 1),
	}
}

func (s *StaticBitModel) SetProbability0(p0 float64) {
	if p0 < 0.0001 || p0 > 0.9999 {
		AC_Error("invalid bit probability")
	}
	s.bit_0_prob = uint32(p0 * (1 << BM__LengthShift))
}
