package guid

import (
	"fmt"
	"math/rand"
	"os"
)

// GUID is a unique identifier designed to virtually guarantee non-conflict between values generated
// across a distributed system.
type GUID struct {
	timeHighAndVersion      uint16
	timeMid                 uint16
	timeLow                 uint32
	clockSeqHighAndReserved uint8
	clockSeqLow             uint8
	node                    [6]byte
}

var (
	emptyGUID GUID
)

// NewGUID generates and returns a new globally unique identifier
func NewGUID() GUID {
	result, err := version4()
	if err != nil {
		panic(err) //Version 4 (psuedo-random GUID) doesn't use anything that could fail.
	}
	return result
}

// Empty returns a copy of the default and empty GUID.
func Empty() GUID {
	return emptyGUID
}

// These constants define the possible string formats available via this implementation of Guid.
const (
	FormatB       string = "B"
	FormatD       string = "D"
	FormatN       string = "N"
	FormatP       string = "P"
	FormatX       string = "X"
	FormatDefault string = FormatD
)

var knownFormats = map[string]string{
	FormatN: "%08x%04x%04x%02x%02x%02x%02x%02x%02x%02x%02x",
	FormatD: "%08x-%04x-%04x-%02x%02x-%02x%02x%02x%02x%02x%02x",
	FormatB: "{%08x-%04x-%04x-%02x%02x-%02x%02x%02x%02x%02x%02x}",
	FormatP: "(%08x-%04x-%04x-%02x%02x-%02x%02x%02x%02x%02x%02x)",
	FormatX: "{0x%08x,0x%04x,0x%04x,{0x%02x,0x%02x,0x%02x,0x%02x,0x%02x,0x%02x,0x%02x,0x%02x}}",
}

// Parse instantiates a GUID from a text represention of the same GUID.
// This is the inverse function of String()
func Parse(value string) (GUID, error) {
	var guid GUID
	for _, fullFormat := range knownFormats {
		parity, err := fmt.Sscanf(
			value,
			fullFormat,
			&guid.timeLow,
			&guid.timeMid,
			&guid.timeHighAndVersion,
			&guid.clockSeqHighAndReserved,
			&guid.clockSeqLow,
			&guid.node[0],
			&guid.node[1],
			&guid.node[2],
			&guid.node[3],
			&guid.node[4],
			&guid.node[5])
		if parity == 11 && err == nil {
			return guid, err
		}
	}
	return emptyGUID, fmt.Errorf("\"%s\" is not in a recognized format", value)
}

func (guid *GUID) String() string {
	result, _ := guid.Format(FormatDefault)
	return result
}

// Format returns a text representation of a GUID that conforms to the specified format.
func (guid *GUID) Format(format string) (string, error) {
	if format == "" {
		format = FormatDefault
	}
	fullFormat, present := knownFormats[format]
	if !present {
		return "", fmt.Errorf("%s is not a recognized format string. Please choose from", format)
	}
	return fmt.Sprintf(
		fullFormat,
		guid.timeLow,
		guid.timeMid,
		guid.timeHighAndVersion,
		guid.clockSeqHighAndReserved,
		guid.clockSeqLow,
		guid.node[0],
		guid.node[1],
		guid.node[2],
		guid.node[3],
		guid.node[4],
		guid.node[5]), nil
}

// Version reads a GUID to parse which mechanism of generating GUIDS was employed.
// Values returned here are documented in rfc4122.txt.
func (guid *GUID) Version() uint {
	return uint(guid.timeHighAndVersion >> 12)
}

const (
	creationFileMode      os.FileMode = os.ModeExclusive
	creationStateFilePath string      = "foo.lock"
)

func version1() (GUID, error) {
	var retval GUID
	return retval, nil
}

func version4() (GUID, error) {
	var retval GUID
	var bits uint32

	// Randomly set all time components and version
	bits = rand.Uint32()
	retval.timeHighAndVersion |= uint16(bits >> 16)
	retval.timeMid |= uint16(bits)
	bits = rand.Uint32()
	retval.timeLow = bits
	bits = rand.Uint32()
	retval.clockSeqHighAndReserved = uint8(bits)
	retval.clockSeqLow = uint8(bits >> 8)

	//Randomly set clock-sequence, reserved, and node
	if written, err := rand.Read(retval.node[:]); !(nil == err && written == len(retval.node)) {
		retval = emptyGUID
		return retval, err
	}

	if err := retval.setVersion(4); nil != err {
		return emptyGUID, err
	}
	retval.clockSeqHighAndReserved = (retval.clockSeqHighAndReserved & 0x3f) | 0x80

	return retval, nil
}

func (guid *GUID) setVersion(version uint16) error {
	if version > 5 {
		return fmt.Errorf("While setting GUID version, unsupported version: %d", version)
	}
	guid.timeHighAndVersion = (guid.timeHighAndVersion & 0x0fff) | version<<12
	return nil
}
