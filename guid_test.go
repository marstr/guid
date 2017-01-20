package guid

import (
	"fmt"
	"testing"
)

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
	subject, err := NewGUIDs(CreationStrategyVersion4)
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
			t.Logf("\nExpected: %s\nActual:  %s", scenario.expected, result)
			t.Fail()
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

func Test_version4_SubsequentCallsDiffer(t *testing.T) {
	seen := make(map[GUID]struct{})
	for i := 0; i < 500; i++ {
		result, _ := version4()
		if _, present := seen[result]; present == true {
			t.Logf("The value %s was generated multiple times.", result.String())
			t.Fail()
		}
		seen[result] = struct{}{}
	}
}

func Test_version1_SubsequentCallsDiffer(t *testing.T) {
	seen := make(map[GUID]struct{})
	for i := 0; i < 500; i++ {
		result, _ := version1()
		if _, present := seen[result]; present == true {
			t.Logf("The value %s was generated multiple times.", result.String())
			t.FailNow()
		}
		seen[result] = struct{}{}
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
	var fodder GUID
	err := fodder.setVersion(0)
	if nil == err {
		t.Log("error expected but unfound when version set to 0")
		t.Fail()
	}

	err = fodder.setVersion(6)
	if nil == err {
		t.Log("error expected but unfound when version set to 6")
		t.Fail()
	}
}

func Benchmark_NewGUID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewGUID()
	}
}

func Benchmark_version1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		version1()
	}
}

func Benchmark_version4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		version4()
	}
}

func ExampleGUID_NewGUIDs() {
	allRandom, err := NewGUIDs(CreationStrategyVersion4)
	if err != nil {
		panic(err)
	}
	fmt.Println(allRandom)
}

func ExampleGUID_Stringf() {
	target := Empty()
	result, _ := target.Stringf(FormatB)
	fmt.Printf(result)
	// Output: {00000000-0000-0000-0000-000000000000}
}
