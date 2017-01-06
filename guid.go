package guid

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
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

// Format enumerates the values that are supported by Parse and Format
type Format string

// These constants define the possible string formats available via this implementation of Guid.
const (
	FormatB       Format = "B"
	FormatD       Format = "D"
	FormatN       Format = "N"
	FormatP       Format = "P"
	FormatX       Format = "X"
	FormatDefault Format = FormatD
)

// CreationStrategy enumerates the values that are supported for populating the bits of a new Guid.
type CreationStrategy string

// These constants define the possible creation strategies available via this implementation of Guid.
const (
	CreationStrategyRFC4122Version1 CreationStrategy = "version1"
	CreationStrategyRFC4122Version2 CreationStrategy = "version2"
	CreationStrategyRFC4122Version3 CreationStrategy = "version3"
	CreationStrategyRFC4122Version4 CreationStrategy = "version4"
	CreationStrategyRFC4122Version5 CreationStrategy = "version5"
)

var emptyGUID GUID

// NewGUID generates and returns a new globally unique identifier
func NewGUID() GUID {
	result, err := version4()
	if err != nil {
		panic(err) //Version 4 (pseudo-random GUID) doesn't use anything that could fail.
	}
	return result
}

var knownStrategies = map[CreationStrategy]func() (GUID, error){
	CreationStrategyRFC4122Version1: version1,
	CreationStrategyRFC4122Version4: version4,
}

// NewGUIDs generates and returns a new globally unique identifier that conforms to the given strategy.
func NewGUIDs(strategy CreationStrategy) (GUID, error) {
	if creator, present := knownStrategies[strategy]; present {
		result, err := creator()
		return result, err
	}
	return emptyGUID, errors.New("Unsupported CreationStrategy")
}

// Empty returns a copy of the default and empty GUID.
func Empty() GUID {
	return emptyGUID
}

var knownFormats = map[Format]string{
	FormatN: "%08x%04x%04x%02x%02x%02x%02x%02x%02x%02x%02x",
	FormatD: "%08x-%04x-%04x-%02x%02x-%02x%02x%02x%02x%02x%02x",
	FormatB: "{%08x-%04x-%04x-%02x%02x-%02x%02x%02x%02x%02x%02x}",
	FormatP: "(%08x-%04x-%04x-%02x%02x-%02x%02x%02x%02x%02x%02x)",
	FormatX: "{0x%08x,0x%04x,0x%04x,{0x%02x,0x%02x,0x%02x,0x%02x,0x%02x,0x%02x,0x%02x,0x%02x}}",
}

// Parse instantiates a GUID from a text representation of the same GUID.
// This is the inverse of function family String()
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

// String returns a text representation of a GUID in the default format.
func (guid *GUID) String() string {
	if result, err := guid.Stringf(FormatDefault); err == nil {
		return result
	}
	return ""
}

// Stringf returns a text representation of a GUID that conforms to the specified format.
func (guid *GUID) Stringf(format Format) (string, error) {
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

var unixToGregorianOffset = time.Date(1970, 01, 01, 0, 0, 00, 0, time.UTC).Sub(time.Date(1582, 10, 15, 0, 0, 0, 0, time.UTC))

// getRFC4122Time returns a 60-bit count of 100-nanosecond intervals since 00:00:00.00 October 15th, 1582
func getRFC4122Time() int64 {
	currentTime := time.Now().UTC().Add(unixToGregorianOffset).UnixNano()
	currentTime /= 100
	return currentTime & 0x0FFFFFFFFFFFFFFF
}

var clockSeqVal uint16
var clockSeqKey sync.Mutex

func getClockSequence() (uint16, error) {
	clockSeqKey.Lock()
	defer clockSeqKey.Unlock()

	if 0 == clockSeqVal {
		var temp [2]byte
		if parity, err := rand.Read(temp[:]); !(2 == parity && nil == err) {
			return 0, err
		}
		clockSeqVal = uint16(temp[0])<<8 | uint16(temp[1])
	}
	clockSeqVal++
	return clockSeqVal, nil
}

func getMACAddress() ([6]byte, error) {
	var retval [6]byte
	var err error
	var hostNICs []net.Interface

	hostNICs, err = net.Interfaces()
	if nil == err {
		found := false
		for _, nic := range hostNICs {
			var parity int

			parity, err = fmt.Sscanf(
				strings.ToLower(nic.HardwareAddr.String()),
				"%02x:%02x:%02x:%02x:%02x:%02x",
				&retval[0],
				&retval[1],
				&retval[2],
				&retval[3],
				&retval[4],
				&retval[5])
			if parity == len(retval) {
				found = true
				break
			}
		}
		if !found {
			err = fmt.Errorf("No suitable address found")
		}
	}
	return retval, err
}

func version1() (GUID, error) {
	var retval GUID
	var localMAC [6]byte
	var err error
	var clockSeq uint16

	currentTime := getRFC4122Time()

	retval.timeLow = uint32(currentTime)
	retval.timeMid = uint16(currentTime >> 32)
	retval.timeHighAndVersion = uint16(currentTime >> 48)
	if err = retval.setVersion(1); err != nil {
		return emptyGUID, err
	}

	if localMAC, err = getMACAddress(); nil != err {
		if parity, err := rand.Read(localMAC[:]); !(len(localMAC) != parity && err == nil) {
			return emptyGUID, err
		}
		localMAC[0] |= 0x1
	}
	copy(retval.node[:], localMAC[:])

	if clockSeq, err = getClockSequence(); nil != err {
		return emptyGUID, err
	}

	retval.clockSeqLow = uint8(clockSeq)
	retval.clockSeqHighAndReserved = uint8(clockSeq >> 8)

	retval.setReservedBits()

	return retval, nil
}

func version4() (GUID, error) {
	var retval GUID
	var bits [10]byte

	if parity, err := rand.Read(bits[:]); !(len(bits) == parity && err == nil) {
		return emptyGUID, err
	}
	retval.timeHighAndVersion |= uint16(bits[0]) | uint16(bits[1])<<8
	retval.timeMid |= uint16(bits[2]) | uint16(bits[3])<<8
	retval.timeLow |= uint32(bits[4]) | uint32(bits[5])<<8 | uint32(bits[6])<<16 | uint32(bits[7])<<24
	retval.clockSeqHighAndReserved = uint8(bits[8])
	retval.clockSeqLow = uint8(bits[9])

	//Randomly set clock-sequence, reserved, and node
	if written, err := rand.Read(retval.node[:]); !(nil == err && written == len(retval.node)) {
		retval = emptyGUID
		return retval, err
	}

	if err := retval.setVersion(4); nil != err {
		return emptyGUID, err
	}
	retval.setReservedBits()

	return retval, nil
}

func (guid *GUID) setVersion(version uint16) error {
	if version > 5 || version == 0 {
		return fmt.Errorf("While setting GUID version, unsupported version: %d", version)
	}
	guid.timeHighAndVersion = (guid.timeHighAndVersion & 0x0fff) | version<<12
	return nil
}

func (guid *GUID) setReservedBits() {
	guid.clockSeqHighAndReserved = (guid.clockSeqHighAndReserved & 0x3f) | 0x80
}
