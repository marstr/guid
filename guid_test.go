package guid

import "testing"

func Test_DefaultIsVersion4(t *testing.T) {
	subject := NewGUID()
	if ver := subject.Version(); ver != 4 {
		t.Errorf("Default GUID should be produced algorithm: version 4. Actual: version %d\n%s", ver, subject.String())
	}
}

func Test_EmptyIsAllZero(t *testing.T) {
	subject := Empty()
	expected := "00000000-0000-0000-0000-000000000000"
	if actual := subject.String(); expected != actual {
		t.Errorf("Empty value expected to be all zero.\nExpected: %s\n  Actual: %s", expected, actual)
	}
}

func Test_Format_Empty(t *testing.T) {
	subject := Empty()
	testCases := []struct {
		shortFormat string
		expected    string
	}{
		{"P", "(00000000-0000-0000-0000-000000000000)"},
		{"X", "{0x00000000,0x0000,0x0000,{0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00}}"},
		{"N", "00000000000000000000000000000000"},
		{"D", "00000000-0000-0000-0000-000000000000"},
		{"B", "{00000000-0000-0000-0000-000000000000}"},
	}

	for _, scenario := range testCases {
		result, err := subject.Format(scenario.shortFormat)
		if nil != err {
			t.Error(err)
		}
		if result != scenario.expected {
			t.Errorf("\nExpected: %s\nActual:  %s", scenario.expected, result)
		}
	}
}

func Test_version4ReservedBits(t *testing.T) {
	for i := 0; i < 500; i++ {
		result, _ := version4()
		if result.clockSeqHighAndReserved&0xc0 != 0x80 {
			t.Fail()
		}
	}
}

func Test_NoOctetisReliablyZero_Default(t *testing.T) {
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
		current := NewGUID()
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

func Test_SubsequentCallsDiffer_Default(t *testing.T) {
	seen := make(map[GUID]struct{})
	for i := 0; i < 500; i++ {
		if i != len(seen) {
			t.Fail()
		}
		result := NewGUID()
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
