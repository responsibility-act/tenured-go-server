package snowflake

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

// These constants are the bit lengths of Snowflake ID parts.
const (
	BitLenTime      = 39                               // bit length of time
	BitLenSequence  = 8                                // bit length of sequence number
	BitLenMachineID = 63 - BitLenTime - BitLenSequence // bit length of machine id

	MaxTime      = uint64((1 << BitLenTime) - 1)
	MaxMachineID = uint16((1 << BitLenMachineID) - 1)
	MaxSequence  = uint8((1 << BitLenSequence) - 1)

	SnowflakeTimeUnit = int64(time.Millisecond)
)

// Settings configures Snowflake:
//
// StartTime is the time since which the Snowflake time is defined as the elapsed time.
// If StartTime is 0, the start time of the Snowflake is set to "2014-09-01 00:00:00 +0000 UTC".
// If StartTime is ahead of the current time, Snowflake is not created.
//
// MachineID returns the unique ID of the Snowflake instance.
// If MachineID returns an error, Snowflake is not created.
// If MachineID is nil, default MachineID is used.
// Default MachineID returns the lower 16 bits of the private IP address.
//
// CheckMachineID validates the uniqueness of the machine ID.
// If CheckMachineID returns false, Snowflake is not created.
// If CheckMachineID is nil, no validation is done.
type Settings struct {
	StartTime time.Time
	MachineID uint16
}

type Petal struct {
	Id        uint64
	Msb       uint8
	Time      uint64
	Sequence  uint8
	MachineId uint16
}

func (p *Petal) String() string {
	return fmt.Sprintf("id: %v, msb: %v, time: %v, sequence: %v, machineId: %v",
		p.Id, p.Msb, p.Time, p.Sequence, p.MachineId)
}

// Snowflake is a distributed unique ID generator.
type Snowflake struct {
	mutex       *sync.Mutex
	startTime   int64
	elapsedTime int64
	sequence    uint16
	machineID   uint16
}

// NewSnowflake returns a new Snowflake configured with the given Settings.
// NewSnowflake returns nil in the following cases:
// - Settings.StartTime is ahead of the current time.
// - Settings.MachineID returns an error.
// - Settings.CheckMachineID returns false.
func NewSnowflake(st Settings) *Snowflake {
	sf := new(Snowflake)
	sf.mutex = new(sync.Mutex)
	sf.sequence = uint16(1<<BitLenSequence - 1)

	if sf.Settings(st) != nil {
		return nil
	}
	return sf
}

func (sf *Snowflake) Settings(st Settings) error {
	if st.StartTime.After(time.Now()) {
		return errors.New("Start time cannot after current time!")
	}
	if st.StartTime.IsZero() {
		sf.startTime = toSnowflakeTime(time.Date(2019, 4, 1, 0, 0, 0, 0, time.UTC))
	} else {
		sf.startTime = toSnowflakeTime(st.StartTime)
	}
	sf.machineID = st.MachineID
	return nil
}

// NextID generates a next unique ID.
// After the Snowflake time overflows, NextID returns an error.
func (sf *Snowflake) NextID() (uint64, error) {
	const maskSequence = uint16(1<<BitLenSequence - 1)

	sf.mutex.Lock()
	defer sf.mutex.Unlock()

	current := currentElapsedTime(sf.startTime)
	if sf.elapsedTime < current {
		sf.elapsedTime = current
		sf.sequence = 0
	} else { // sf.elapsedTime >= current
		sf.sequence = (sf.sequence + 1) & maskSequence
		if sf.sequence == 0 {
			sf.elapsedTime++
			overtime := sf.elapsedTime - current
			time.Sleep(sleepTime(overtime))
		}
	}

	return sf.toID()
}

func toSnowflakeTime(t time.Time) int64 {
	return t.UTC().UnixNano() / SnowflakeTimeUnit
}

func currentElapsedTime(startTime int64) int64 {
	return toSnowflakeTime(time.Now()) - startTime
}

func sleepTime(overtime int64) time.Duration {
	return time.Duration(overtime)*10*time.Millisecond -
		time.Duration(time.Now().UTC().UnixNano()%SnowflakeTimeUnit)*time.Nanosecond
}

func (sf *Snowflake) toID() (uint64, error) {
	if sf.elapsedTime >= 1<<BitLenTime {
		return 0, errors.New("over the time limit")
	}

	return uint64(sf.elapsedTime)<<(BitLenSequence+BitLenMachineID) |
		uint64(sf.sequence)<<BitLenMachineID |
		uint64(sf.machineID), nil
}

func privateIPv4() (net.IP, error) {
	as, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, a := range as {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}

		ip := ipnet.IP.To4()
		if isPrivateIPv4(ip) {
			return ip, nil
		}
	}
	return nil, errors.New("no private ip address")
}

func isPrivateIPv4(ip net.IP) bool {
	return ip != nil &&
		(ip[0] == 10 || ip[0] == 172 && (ip[1] >= 16 && ip[1] < 32) || ip[0] == 192 && ip[1] == 168)
}

func lower16BitPrivateIP() (uint16, error) {
	ip, err := privateIPv4()
	if err != nil {
		return 0, err
	}
	return uint16(ip[2])<<8 + uint16(ip[3]), nil
}

// Decompose returns a set of Snowflake ID parts.
func Decompose(id uint64) *Petal {
	const maskSequence = uint64((1<<BitLenSequence - 1) << BitLenMachineID)
	const maskMachineID = uint64(1<<BitLenMachineID - 1)

	msb := id >> 63
	time := id >> (BitLenSequence + BitLenMachineID)
	sequence := id & maskSequence >> BitLenMachineID
	machineID := id & maskMachineID
	return &Petal{
		Id:        id,
		Msb:       uint8(msb),
		Time:      time,
		Sequence:  uint8(sequence),
		MachineId: uint16(machineID),
	}
}
