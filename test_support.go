package FastAC

import (
	"fmt"
	"math"
	"time"
)

const MinProbability = 1e-4 // 0.0001

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

type Chronometer struct {
	on   bool
	mark time.Time
	time time.Duration
}

func (c *Chronometer) Reset() {
	c.time = 0
	c.on = false
}

func (c *Chronometer) Start(s string) {
	if s != "" {
		fmt.Println(s)
	}
	if c.on {
		fmt.Println("chronometer already on!")
	} else {
		c.on = true
		c.mark = time.Now()
	}
}

func (c *Chronometer) Restart(s string) {
	if s != "" {
		fmt.Println(s)
	}
	c.time = 0
	c.on = true
	c.mark = time.Now()
}

func (c *Chronometer) Stop() {
	if c.on {
		c.on = false
		c.time = time.Since(c.mark)
	} else {
		fmt.Println("chronometer already off!")
	}
}

func (c *Chronometer) Read() time.Duration {
	if c.on {
		return c.time + time.Since(c.mark)
	} else {
		return c.time
	}
}

func (c *Chronometer) Display(s string) {
	var sc time.Duration
	if c.on {
		sc = c.time + time.Since(c.mark)
	} else {
		sc = c.time
	}
	// pretty print seconds
	fmt.Printf("%s: %5.2fs\n", s, float64(sc)/float64(time.Second))
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

type RandomGenerator struct {
	s1, s2, s3 uint32
}

func initRandomGenerator(seed uint32) *RandomGenerator {
	rg := new(RandomGenerator)
	rg.SetSeed(seed)
	return rg
}

func (rg *RandomGenerator) SetSeed(seed uint32) {
	rg.s1 = 0x147AE11
	if seed != 0 {
		rg.s1 = seed & 0xFFFFFFF
	}
	rg.s2 = rg.s1 ^ 0xFFFFF07
	rg.s3 = rg.s1 ^ 0xF03CD2F
}

func (rg *RandomGenerator) Word() uint32 {
	// "Taus88" generator with period (2^31 - 1) * (2^29 - 1) * (2^28 - 1)
	// FastAC/test/test_support.h

	var b uint32
	b = ((rg.s1 << 13) ^ rg.s1) >> 19
	rg.s1 = ((rg.s1 & 0xFFFFFFFE) << 12) ^ b
	b = ((rg.s2 << 2) ^ rg.s2) >> 25
	rg.s2 = ((rg.s2 & 0xFFFFFFF8) << 4) ^ b
	b = ((rg.s3 << 3) ^ rg.s3) >> 11
	rg.s3 = ((rg.s3 & 0xFFFFFFF0) << 17) ^ b
	return rg.s1 ^ rg.s2 ^ rg.s3
}

func (rg *RandomGenerator) Uniform() float64 {
	const WordToDouble = 1.0 / (1.0 + float64(0xFFFFFFFF))

	var b uint32
	b = ((rg.s1 << 13) ^ rg.s1)   >> 19
	rg.s1 = ((rg.s1 & 0xFFFFFFFE) << 12) ^ b
	b = ((rg.s2 << 2) ^ rg.s2)    >> 25
	rg.s2 = ((rg.s2 & 0xFFFFFFF8) << 4)  ^ b
	b = ((rg.s3 << 3) ^ rg.s3)    >> 11
	rg.s3 = ((rg.s3 & 0xFFFFFFF0) << 17) ^ b // open interval: 0 < r < 1
	return WordToDouble * (0.5 + float64(rg.s1^rg.s2^rg.s3))
}

func (rg *RandomGenerator) Integer(Range uint32) uint32 {
	return uint32(float64(Range) * rg.Uniform())
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

type RandomBitSource struct {
	*RandomGenerator // RandomBitSource implements RandomGenerator
	threshold        uint32
	ent, prob0       float64
}

func (rbs *RandomBitSource) entropy() float64 {
	return rbs.ent
}

func (rbs *RandomBitSource) symbol0probability() float64 {
	return rbs.prob0
}

func (rbs *RandomBitSource) symbol1probability() float64 {
	return 1.0 - rbs.prob0
}

func initRandomBitSource() *RandomBitSource {
	rbs := new(RandomBitSource)
	rbs.RandomGenerator = initRandomGenerator(0)
	rbs.prob0 = 0.5
	rbs.ent = 1.0
	return rbs
}

func (rbs *RandomBitSource) SetProbability0(p0 float64) float64 {
	if p0 < MinProbability || p0 > 1.0-MinProbability {
		AC_Error("invalid random bit probability")
	}

	rbs.prob0 = p0
	rbs.threshold = uint32(p0 * 0xFFFFFFFF)
	rbs.ent = ((p0-1.0)*math.Log(1.0-p0) - p0*math.Log(p0)) / math.Log(2.0)

	return rbs.ent
}

func (rbs *RandomBitSource) SetEntropy(entropy float64) {
	if entropy < 0.0001 || entropy > 1.0 {
		AC_Error("invalid random bit entropy")
	}

	var h, p float64 = entropy * math.Log(2.0), 0.5 * entropy * entropy
	for k := 0; k < 8; k++ {
		var lp1 float64 = math.Log(1.0 - p)
		var lp2 float64 = lp1 - math.Log(p)
		var d float64 = h + lp1 - p*lp2
		if math.Abs(d) < 1e-12 {
			break
		}
		p += d / lp2
	}
	rbs.SetProbability0(p)
}

func (rbs *RandomBitSource) SwitchProbabilities() {
	rbs.SetProbability0(1.0 - rbs.prob0)
}

func (rbs *RandomBitSource) ShuffleProbabilities() {
	if rbs.Word() > 0x80000000 {
		rbs.SetProbability0(1.0 - rbs.prob0)
	}
}

func (rbs *RandomBitSource) Bit() uint {
	if rbs.Word() > rbs.threshold {
		return 1
	}
	return 0
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

type RandomDataSource struct {
	*RandomGenerator // RandomDataSource implements RandomGenerator
	ent              float64
	prob             []float64

	symbols  uint32
	dist     []uint32
	lowBound [257]uint32
}

func (rds *RandomDataSource) entropy() float64 {
	return rds.ent
}

func (rds *RandomDataSource) probability() []float64 {
	return rds.prob
}

func (rds *RandomDataSource) data_symbols() float64 {
	return float64(rds.symbols)
}

func initRandomDataSource() *RandomDataSource {
	rds := new(RandomDataSource) // symbols dist and prob must be 0
	rds.RandomGenerator = initRandomGenerator(0)
	return rds
}

func (rds *RandomDataSource) AssignMemory(dim uint32) {
	if rds.symbols == dim {
		return
	}
	rds.symbols = dim
	rds.dist = nil
	rds.prob = nil
	rds.prob = make([]float64, dim)
	rds.dist = make([]uint32, dim)
}

func (rds *RandomDataSource) SetDistribution(dim uint32, probability []float64) float64 {
	rds.AssignMemory(dim)

	var sum float64 = 0
	rds.ent = 0

	var s uint32 = 0
	rds.lowBound[0] = 0

	const DoubleToWord float64 = 1.0 + float64(0xFFFFFFFF)

	for n := uint32(0); n < rds.symbols; n++ {
		p := probability[n]
		if p < MinProbability {
			AC_Error("invalid random source probability")
		}
		rds.prob[n] = p
		rds.dist[n] = uint32(0.49 + DoubleToWord*sum)
		w := rds.dist[n] >> 24
		for s < w {
			s++
			rds.lowBound[s] = n - 1
		}
		sum += p
		rds.ent -= p * math.Log(p)
	}

	for s < 256 {
		s++
		rds.lowBound[s] = rds.symbols - 1
	}

	if math.Abs(1.0-sum) > 1e-4 {
		AC_Error("invalid random source distribution")
	}
	rds.ent /= math.Log(2.0)
	return rds.ent
}

func (rds *RandomDataSource) SetTG(a float64) float64 {
	var s, r, e, m float64 = 0, 0, 0, float64(rds.symbols)

	if a > 1e-4 {
		s = (1.0 - math.Exp(-a)) / (1.0 - math.Exp(-a*m))
	} else {
		s = (2.0 - a) / (m * (2.0 - a*m))
	}

	for n := int(rds.symbols - 1); n >= 0; n-- {
		var p float64
		if a*float64(n) > 30.0 {
			p = 0
		} else {
			p = s * math.Exp(-a*float64(n))
		}

		if p < MinProbability {
			r += MinProbability - p
			p = MinProbability
		} else {
			if r > 0 {
				if r <= p-MinProbability {
					p -= r
					r = 0
				} else {
					r -= p - MinProbability
					p = MinProbability
				}
			}
		}
		rds.prob[n] = p
		e -= p * math.Log(p)
	}

	return e / math.Log(2.0)
}

func (rds *RandomDataSource) SetTruncatedGeometric(dim uint32, entropy float64) float64 {
	rds.AssignMemory(dim)

	max_entropy := math.Log(float64(rds.symbols)) / math.Log(2.0)
	mgr_prob := float64(dim-1) * MinProbability
	min_entropy := ((mgr_prob-1.0)*math.Log(1.0-mgr_prob) - mgr_prob*math.Log(MinProbability)) * 1.2 / math.Log(2.0)

	if entropy <= min_entropy || entropy > max_entropy {
		AC_Error("invalid data source entropy")
	}

	ZF := initZeroFinder(0, 2)
	a := ZF.SetNewResult(max_entropy - entropy)

	for itr := uint(0); itr < 20; itr++ {
		ne := rds.SetTG(a) - entropy
		if math.Abs(ne) < 1e-5 {
			break
		}
		a = ZF.SetNewResult(ne)
	}

	rds.SetDistribution(rds.symbols, rds.prob)
	if math.Abs(rds.ent-entropy) > 1e-4 {
		AC_Error("cannot set random source entropy")
	}

	return rds.ent
}

func (rds *RandomDataSource) ShuffleProbabilities() {
	for n := rds.symbols - 1; n > 0; n-- {
		m := rds.Integer(n + 1)
		if m == n {
			continue
		}
		t := rds.prob[m]
		rds.prob[m] = rds.prob[n]
		rds.prob[n] = t
	}
	rds.SetDistribution(rds.symbols, rds.prob)
}

func (rds *RandomDataSource) Data() uint32 {
	v := rds.Word()
	w := v >> 24
	u, n := rds.lowBound[w], rds.lowBound[w+1]+1
	for n > u+1 {
		m := (u + n) >> 1
		if rds.dist[m] < v {
			u = m
		} else {
			n = m
		}
	}
	return u
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

type ZeroFinder struct {
	phase, iter int

	x0, y0, x1, y1, x2, y2, x float64
}

func initZeroFinder(first_x, second_x float64) *ZeroFinder {
	zf := new(ZeroFinder)
	zf.x0 = first_x
	zf.x1 = second_x
	return zf
}

func (zf *ZeroFinder) SetNewResult(y float64) float64 {
	zf.iter++
	if zf.iter > 30 {
		AC_Error("cannot find solution")
	}

	if zf.phase >= 2 {
		if y*zf.y0 <= 0 {
			if (zf.phase == 2) || math.Abs(zf.y1) < math.Abs(zf.y2) {
				zf.x2 = zf.x1
				zf.y2 = zf.y1
			}
			zf.x1 = zf.x
			zf.y1 = y
		} else {
			if zf.phase == 2 || math.Abs(zf.y0) < math.Abs(zf.y2) {
				zf.x2 = zf.x0
				zf.y2 = zf.y0
			}
			zf.x0 = zf.x
			zf.y0 = y
		}

		if math.Abs(zf.y0) < math.Abs(zf.y1) {
			r, c := zf.y0/zf.y2, zf.x2-zf.x0
			s, d := zf.y0/zf.y1, zf.x1-zf.x0
			zf.x = zf.x0 - (c*d*(s-r))/(c*(1.0-s)-d*(1.0-r))
		} else {
			r, c := zf.y1/zf.y2, zf.x2-zf.x1
			s, d := zf.y1/zf.y0, zf.x0-zf.x1
			zf.x = zf.x1 - (c*d*(s-r))/(c*(1.0-s)-d*(1.0-r))
		}
		zf.phase = 3
		return zf.x
	}

	if zf.iter > 8 {
		AC_Error("too many initial tests")
	}

	if zf.phase == 1 {
		if y*zf.y0 <= 0 {
			zf.y1 = y
			zf.phase = 2
			if math.Abs(zf.y0) < math.Abs(zf.y1) {
				s := zf.y0 / zf.y1
				zf.x = zf.x0 - ((zf.x1-zf.x0)*s)/(1.0-s)
			} else {
				s := zf.y1 / zf.y0
				zf.x = zf.x1 - ((zf.x0-zf.x1)*s)/(1.0-s)
			}
		} else {
			zf.x += zf.x1 - zf.x0
			zf.x0 = zf.x1
			zf.y0 = y
			zf.x1 = zf.x
		}
		return zf.x
	}
	zf.y0 = y
	zf.phase = 1

	zf.x = zf.x1
	return zf.x
}
