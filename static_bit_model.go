package FastAC

type StaticBitModel struct {
	bit0prob uint32
}

func NewStaticBitModel() *StaticBitModel {
	return &StaticBitModel{
		bit0prob: 1 << (BM__LengthShift - 1),
	}
}

func (s *StaticBitModel) SetProbability0(p0 float64) {
	if p0 < 0.0001 || p0 > 0.999 {
		AC_Error("invalid bit probability")
	}
	s.bit0prob = uint32(p0 * (1 << BM__LengthShift))
}
