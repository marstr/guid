package guid

import (
	"fmt"
	"testing"
)

func Test_DefaultIsVersion4(t *testing.T) {
	subject := NewGUID()
	if ver := subject.Version(); ver != 4 {
		t.Logf("Default GUID should be produced using algorithm: version 4. Actual: version %d\n%s", ver, subject.String())
		t.Fail()
	}
}

func Test_NewGUIDs_NotEmpty(t *testing.T) {
	for strat := range knownStrategies {
		t.Run(string(strat), func(subT *testing.T) {
			subject, err := NewGUIDs(strat)
			if err != nil {
				subT.Error(err)
			}
			if subject == Empty() {
				subT.Logf("unexpected empty encountered")
				subT.Fail()
			}
		})
	}
}

func Test_NewGUIDs_Unsupported(t *testing.T) {
	fauxStrategy := CreationStrategy("invalidStrategy")
	subject, err := NewGUIDs(fauxStrategy)
	if subject != Empty() {
		t.Fail()
	}
	if err.Error() != "Unsupported CreationStrategy" {
		t.Fail()
	}
}

func Test_Format_Empty(t *testing.T) {
	subject := Empty()
	testCases := []struct {
		shortFormat Format
		expected    string
	}{
		{FormatP, "(00000000-0000-0000-0000-000000000000)"},
		{FormatX, "{0x00000000,0x0000,0x0000,{0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00}}"},
		{FormatN, "00000000000000000000000000000000"},
		{FormatD, "00000000-0000-0000-0000-000000000000"},
		{FormatB, "{00000000-0000-0000-0000-000000000000}"},
	}

	for _, scenario := range testCases {
		t.Run("", func(subT *testing.T) {
			result := subject.Stringf(scenario.shortFormat)
			if result != scenario.expected {
				subT.Logf("\nwant:\t%s\ngot: \t%s", scenario.expected, result)
				subT.Fail()
			}
		})
	}
}

func Test_Parse_Roundtrip(t *testing.T) {
	subject := NewGUID()

	for format := range knownFormats {
		t.Run(string(format), func(subT *testing.T) {
			serialized := subject.Stringf(format)

			parsed, parseErr := Parse(serialized)
			if nil != parseErr {
				subT.Error(parseErr)
			}

			if parsed != subject {
				subT.Logf("\nwant:\t%s\ngot: \t%s", subject.String(), parsed.String())
				subT.Fail()
			}
		})
	}
}

func Test_version4_ReservedBits(t *testing.T) {
	for i := 0; i < 500; i++ {
		result, _ := version4()
		if result.clockSeqHighAndReserved&0xc0 != 0x80 {
			t.Fail()
		}
	}
}

func Test__version4_NoOctetisReliablyZero(t *testing.T) {
	results := make(map[string]uint)

	const iterations uint = 500
	const suspicionThreshold uint = iterations / 10

	results["time_low"] = 0
	results["time_mid"] = 0
	results["time_hi_and_version"] = 0
	results["clock_seq_hi_and_reserved"] = 0
	results["clock_seq_low"] = 0
	results["node"] = 0

	for i := uint(0); i < iterations; i++ {
		current, _ := version4()
		if 0 == current.timeLow {
			results["time_low"]++
		}
		if 0 == current.timeMid {
			results["time_mid"]++
		}
		if 0 == current.timeHighAndVersion {
			results["time_hi_and_version"]++
		}
		if 0 == current.clockSeqHighAndReserved {
			results["clock_seq_hi_and_reserved"]++
		}
		if 0 == current.clockSeqLow {
			results["clock_seq_low"]++
		}
	}

	anySuspicious := false
	for key, val := range results {
		if val > suspicionThreshold {
			anySuspicious = true
			t.Logf("%s reported value 0 enough times (%d of %d) to be suspicious.", key, val, iterations)
		}
	}
	if anySuspicious {
		t.Fail()
	}
}

func Test_SubsequentCallsDiffer(t *testing.T) {
	for strat := range knownStrategies {
		t.Run(string(strat), func(subT *testing.T) {
			seen := make(map[GUID]struct{})
			for i := 0; i < 500; i++ {
				result, err := NewGUIDs(strat)
				if err != nil {
					subT.Error(err)
				}
				if _, present := seen[result]; present == true {
					subT.Logf("The value %s was generated multiple times.", result.String())
					subT.Fail()
				}
				seen[result] = struct{}{}
			}
		})
	}

}

func Test_getMACAddress(t *testing.T) {
	subject, err := getMACAddress()
	t.Logf("MAC returned: %02x:%02x:%02x:%02x:%02x:%02x", subject[0], subject[1], subject[2], subject[3], subject[4], subject[5])

	if nil != err {
		t.Error(err)
	}

	nonZeroSeen := false
	for _, octet := range subject {
		if 0 != octet {
			nonZeroSeen = true
			break
		}
	}
	if !nonZeroSeen {
		t.Fail()
	}
}

func Test_setVersion_bounds(t *testing.T) {
	testCases := []uint16{0, 6}
	for _, tc := range testCases {
		t.Run(fmt.Sprint(tc), func(t *testing.T) {
			var fodder GUID
			err := fodder.setVersion(tc)
			if nil == err {
				t.Log("error expected but unfound when version set to 0")
				t.Fail()
			}
		})
	}
}

func Benchmark_NewGUIDs(b *testing.B) {
	for strat := range knownStrategies {
		b.Run(string(strat), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				NewGUIDs(strat)
			}
		})
	}
}

func Benchmark_String(b *testing.B) {
	rand, _ := NewGUIDs(CreationStrategyVersion4)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rand.String()
	}
}

func Benchmark_Stringf(b *testing.B) {
	rand := NewGUID()
	b.ResetTimer()

	for format := range knownFormats {
		b.Run(string(format), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				rand.Stringf(format)
			}
		})
	}
}

func Benchmark_Parse(b *testing.B) {
	rand := NewGUID()
	b.ResetTimer()

	for format := range knownFormats {
		printed := rand.Stringf(format)
		b.Run(string(format), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				Parse(printed)
			}
		})
	}
}

func ExampleGUID_Stringf() {
	fmt.Printf(Empty().Stringf(FormatB))
	// Output: {00000000-0000-0000-0000-000000000000}
}

func ExampleGUID_String() {
	fmt.Printf(Empty().String())
	// Output: 00000000-0000-0000-0000-000000000000
}

func Example_Empty() {
	var example GUID
	if example == Empty() {
		fmt.Print("Example is Empty")
	} else {
		fmt.Print("Example is not Empty")
	}
	// Output: Example is Empty
}
