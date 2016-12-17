package guid

import "testing"

func Test_DefaultIsVersion4(t *testing.T) {
	subject := NewGUID()
	if ver := subject.Version(); ver != 4 {
		t.Errorf("Default GUID should be produced algorithm: version 4. Actual: version %d\n%s", ver, subject.String())
	}
}

func Test_Empty_IsAllZero(t *testing.T) {
	subject := Empty()
	expected := "00000000-0000-0000-0000-000000000000"
	if actual := subject.String(); expected != actual {
		t.Errorf("Empty value expected to be all zero.\nExpected: %s\n  Actual: %s", expected, actual)
	}
}

func Test_NewGUIDs_NotEmpty(t *testing.T) {
	subject, err := NewGUIDs(CreationStrategyRFC4122Version4)
	if err != nil {
		t.Error(err)
	}
	if subject == Empty() {
		t.Error()
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
		result, err := subject.Stringf(scenario.shortFormat)
		if nil != err {
			t.Error(err)
		}
		if result != scenario.expected {
			t.Errorf("\nExpected: %s\nActual:  %s", scenario.expected, result)
		}
	}
}

func Test_Parse_Roundtrip(t *testing.T) {
	subject := NewGUID()

	for format := range knownFormats {
		serialized, formatErr := subject.Stringf(format)
		if nil != formatErr {
			t.Error(formatErr)
		}

		parsed, parseErr := Parse(serialized)
		if nil != parseErr {
			t.Error(parseErr)
		}

		if parsed != subject {
			t.Logf("Expected: %s Actual: %s", subject.String(), parsed.String())
			t.Fail()
		}
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

	iterations := uint(500)
	suspicionThreshold := iterations / 10

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

	for key, val := range results {
		if val > suspicionThreshold {
			t.Errorf("%s reported value 0 enough times (%d of %d) to be suspicious.", key, val, iterations)
		}
	}
}

func Test_version4_SubsequentCallsDiffer(t *testing.T) {
	seen := make(map[GUID]struct{})
	for i := 0; i < 500; i++ {
		if i != len(seen) {
			t.Fail()
		}
		result, _ := version4()
		if _, present := seen[result]; present == true {
			t.Errorf("The value %s was generated multiple times.", result)
			break
		}
		seen[result] = struct{}{}
	}
}

func Benchmark_NewGUID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewGUID()
	}
}
