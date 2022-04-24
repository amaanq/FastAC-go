package FastAC

type Model interface {
	Encode(value uint)
	Decode() uint
}
