package FastAC

import (
	"fmt"
	"math"
	"testing"
)

const SimulTests = 1000000

type TestResult struct {
	alphabetSymbols                uint32
	encoderTime, decoderTime       float64
	entropy, bitsUsed, testSymbols float64
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
// - - Implementations for testing encoder/decoder - - - - - - - - - - - - - -

func EncodeStaticBitBuffer(bitBuffer []byte, model *StaticBitModel, encoder *ArithmeticCodec) uint32 {
	encoder.StartEncoder()
	for k := 0; k < SimulTests; k++ {
		encoder.EncodeFromStaticBitModel(uint32(bitBuffer[k]), model)
	}
	return 8 * encoder.StopEncoder()
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func DecodeStaticBitBuffer(bitBuffer []byte, model *StaticBitModel, decoder *ArithmeticCodec) {
	decoder.StartDecoder()
	for k := 0; k < SimulTests; k++ {
		bitBuffer[k] = byte(decoder.DecodeFromStaticBitModel(model))
	}
	decoder.StopDecoder()
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func EncodeAdaptiveBitBuffer(bitBuffer []byte, model *AdaptiveBitModel, encoder *ArithmeticCodec) uint32 {
	encoder.StartEncoder()
	for k := 0; k < SimulTests; k++ {
		encoder.EncodeFromAdaptiveBitModel(uint32(bitBuffer[k]), model)
	}
	return 8 * encoder.StopEncoder()
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func DecodeAdaptiveBitBuffer(bitBuffer []byte, model *AdaptiveBitModel, decoder *ArithmeticCodec) {
	decoder.StartDecoder()
	for k := 0; k < SimulTests; k++ {
		bitBuffer[k] = byte(decoder.DecodeFromAdaptiveBitModel(model))
	}
	decoder.StopDecoder()
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func EncodeStaticDataBuffer(dataBuffer []uint16, model *StaticDataModel, encoder *ArithmeticCodec) uint32 {
	encoder.StartEncoder()
	for k := 0; k < SimulTests; k++ {
		encoder.EncodeFromStaticDataModel(uint32(dataBuffer[k]), model)
	}
	return 8 * encoder.StopEncoder()
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func DecodeStaticDataBuffer(dataBuffer []uint16, model *StaticDataModel, decoder *ArithmeticCodec) {
	decoder.StartDecoder()
	for k := 0; k < SimulTests; k++ {
		dataBuffer[k] = uint16(decoder.DecodeFromStaticDataModel(model))
	}
	decoder.StopDecoder()
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func EncodeAdaptiveDataBuffer(dataBuffer []uint16, model *AdaptiveDataModel, encoder *ArithmeticCodec) uint32 {
	encoder.StartEncoder()
	for k := 0; k < SimulTests; k++ {
		encoder.EncodeFromAdaptiveDataModel(uint32(dataBuffer[k]), model)
	}
	return 8 * encoder.StopEncoder()
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func DecodeAdaptiveDataBuffer(dataBuffer []uint16, model *AdaptiveDataModel, decoder *ArithmeticCodec) {
	decoder.StartDecoder()
	for k := 0; k < SimulTests; k++ {
		dataBuffer[k] = uint16(decoder.DecodeFromAdaptiveDataModel(model))
	}
	decoder.StopDecoder()
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func FillBitBuffer(src *RandomBitSource, bitBuffer []byte) {
	src.ShuffleProbabilities()
	for k := 0; k < SimulTests; k++ {
		bitBuffer[k] = byte(src.Bit())
	}
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func FillDataBuffer(src *RandomDataSource, dataBuffer []uint16) {
	src.ShuffleProbabilities()
	for k := 0; k < SimulTests; k++ {
		dataBuffer[k] = uint16(src.Data())
	}
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func DisplayResults(first, adaptive bool, pr *TestResult, sourceTime float64) {
	if adaptive {
		fmt.Println(" Test with adaptive model")
	} else {
		if first {
			fmt.Println("\n=========================================================================")
		}
		fmt.Printf(" Test with static model\n\n")
	}

	fmt.Printf(" Random  data generated in %5.2f seconds\n", sourceTime)
	fmt.Printf(" Encoder test completed in %5.2f seconds\n", pr.encoderTime)
	fmt.Printf(" Decoder test completed in %5.2f seconds\n", pr.decoderTime)

	fmt.Printf("\n Used %g bytes for coding %g symbols\n",
		0.125*pr.bitsUsed, pr.testSymbols)
	fmt.Printf(" Data source entropy = %8.5f bits/symbol [%d symbols]\n",
		pr.entropy, pr.alphabetSymbols)

	bitRate := pr.bitsUsed / pr.testSymbols
	fmt.Printf(" Compression rate    = %8.5f bits/symbol (%9.4f %% redundancy)\n\n",
		bitRate, 100.0*(bitRate-pr.entropy)/pr.entropy)

	fmt.Printf(" Encoding time  = %8.3f ns/symbol  = %8.3f ns/bit\n",
		1e9*pr.encoderTime/pr.testSymbols,
		1e9*pr.encoderTime/pr.bitsUsed)

	fmt.Printf(" Decoding time  = %8.3f ns/symbol  = %8.3f ns/bit\n",
		1e9*pr.decoderTime/pr.testSymbols,
		1e9*pr.decoderTime/pr.bitsUsed)

	fmt.Printf(" Encoding speed = %8.3f Msymbols/s = %8.3f Mbits/s\n",
		1e-6*pr.testSymbols/pr.encoderTime,
		1e-6*pr.bitsUsed/pr.encoderTime)

	fmt.Printf(" Decoding speed = %8.3f Msymbols/s = %8.3f Mbits/s\n",
		1e-6*pr.testSymbols/pr.decoderTime,
		1e-6*pr.bitsUsed/pr.decoderTime)

	fmt.Println("=========================================================================")
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func BinaryBenchmark(num_cycles int) {
	num_simulations := 10
	entropy, entropy_increment := 0.1, 0.1

	result := new(TestResult)
	src := initRandomBitSource()
	codec := NewArithmeticCodec(SimulTests>>2, nil)
	static_model := NewStaticBitModel()
	adaptive_model := NewAdaptiveBitModel()
	encoder_time, decoder_time, source_time := new(Chronometer), new(Chronometer), new(Chronometer)

	var code_bits uint32
	source_bits := make([]byte, 2*SimulTests)
	decoded_bits := make([]byte, 2*SimulTests)

	for simul := 0; simul < num_simulations; simul++ {
		for pass := 0; pass <= 1; pass++ {
			src.SetEntropy(entropy)
			src.SetSeed(1839304 + 2017*uint32(simul))

			result.alphabetSymbols = 2
			result.entropy = src.entropy()
			result.testSymbols = 0
			result.bitsUsed = 0

			source_time.Reset()
			encoder_time.Reset()
			decoder_time.Reset()

			for cycle := 0; cycle < num_cycles; cycle++ {
				source_time.Start("")
				FillBitBuffer(src, source_bits)
				source_time.Stop()

				if pass == 0 {
					static_model.SetProbability0(src.symbol0probability())
					encoder_time.Start("")
					code_bits = EncodeStaticBitBuffer(source_bits, static_model, codec)
					encoder_time.Stop()

					decoder_time.Start("")
					DecodeStaticBitBuffer(decoded_bits, static_model, codec)
					decoder_time.Stop()
				} else {
					adaptive_model.reset()
					encoder_time.Start("")
					code_bits = EncodeAdaptiveBitBuffer(source_bits, adaptive_model, codec)
					encoder_time.Stop()

					adaptive_model.reset()
					decoder_time.Start("")
					DecodeAdaptiveBitBuffer(decoded_bits, adaptive_model, codec)
					decoder_time.Stop()
				}

				result.testSymbols += float64(SimulTests)
				result.bitsUsed += float64(code_bits)

				for k := 0; k < SimulTests; k++ {
					if source_bits[k] != decoded_bits[k] {
						AC_Error("incorrect decoding")
					}
				}
			}

			result.encoderTime = encoder_time.Read().Seconds()
			result.decoderTime = decoder_time.Read().Seconds()
			DisplayResults(simul == 0, pass != 0, result, source_time.Read().Seconds())
		}
		entropy += entropy_increment
	}
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func GeneralBenchmark(data_symbols, num_cycles uint32) {
	var entropy, entropy_increment float64
	if data_symbols <= 8 {
		entropy = 0.2
		entropy_increment = 0.20
	} else {
		if data_symbols <= 32 {
			entropy = 0.5
			entropy_increment = 0.25
		} else {
			entropy = 1.0
			entropy_increment = 0.50
		}
	}

	num_simulations := int(1 + ((math.Log(float64(data_symbols))/math.Log(2.0) - entropy) / entropy_increment))

	result := new(TestResult)
	src := initRandomDataSource()
	codec := NewArithmeticCodec(SimulTests<<1, nil)
	static_model := NewStaticDataModel()
	adaptive_model := NewAdaptiveDataModel(data_symbols)
	encoder_time, decoder_time, source_time := new(Chronometer), new(Chronometer), new(Chronometer)

	var code_bits uint32
	source_data := make([]uint16, SimulTests) // 2000000 bytes is 1000000 uint16s
	decoded_data := make([]uint16, SimulTests)

	adaptive_model.SetAlphabet(data_symbols)

	for simul := 0; simul < int(num_simulations); simul++ {
		for pass := 0; pass <= 1; pass++ {
			src.SetTruncatedGeometric(data_symbols, entropy)
			src.SetSeed(8315739 + 1031*uint32(simul) + 11*uint32(data_symbols))

			result.alphabetSymbols = uint32(data_symbols)
			result.entropy = src.entropy()
			result.testSymbols = 0
			result.bitsUsed = 0

			source_time.Reset()
			encoder_time.Reset()
			decoder_time.Reset()

			for cycle := 0; cycle < int(num_cycles); cycle++ {
				source_time.Start("")
				FillDataBuffer(src, source_data)
				source_time.Stop()

				if pass == 0 {
					static_model.SetDistribution(data_symbols, src.probability())
					encoder_time.Start("")
					code_bits = EncodeStaticDataBuffer(source_data, static_model, codec)
					encoder_time.Stop()

					decoder_time.Start("")
					DecodeStaticDataBuffer(decoded_data, static_model, codec)
					decoder_time.Stop()
				} else {
					adaptive_model.Reset()
					encoder_time.Start("")
					code_bits = EncodeAdaptiveDataBuffer(source_data, adaptive_model, codec)
					encoder_time.Stop()

					adaptive_model.Reset()
					decoder_time.Start("")
					DecodeAdaptiveDataBuffer(decoded_data, adaptive_model, codec)
					decoder_time.Stop()
				}

				result.testSymbols += float64(SimulTests)
				result.bitsUsed += float64(code_bits)

				for k := 0; k < SimulTests; k++ {
					if source_data[k] != decoded_data[k] {
						AC_Error("incorrect decoding")
					}
				}
			}

			result.encoderTime = encoder_time.Read().Seconds()
			result.decoderTime = decoder_time.Read().Seconds()
			DisplayResults(simul == 0, pass != 0, result, source_time.Read().Seconds())
		}
		entropy += entropy_increment
	}
}

func Test(t *testing.T) {
	num_symbols := 3
	total_cycles := 10

	if num_symbols == 2 {
		BinaryBenchmark(num_symbols)
	} else {
		GeneralBenchmark(uint32(num_symbols), uint32(total_cycles))
	}
}
