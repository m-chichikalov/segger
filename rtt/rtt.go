package rtt

import (
	"errors"
)

const (
	MaxNumUpBuffers   = 1
	MaxNumDownBuffers = 1
	SizeUpBuffer      = 1024
	SizeDownBuffer    = 32
)

type BufferUpRTT struct {
	sName        *uint8
	pBuffer      *uint8
	sizeOfBuffer uint32
	wrOff        uint32
	rdOff        uint32 // volatile.Register32
	flags        uint32
}

type BufferDownRTT struct {
	sName        *uint8
	pBuffer      *uint8
	sizeOfBuffer uint32
	wrOff        uint32 // volatile.Register32
	rdOff        uint32
	flags        uint32
}

type ControlBlockRTT struct {
	// Initialized to "SEGGER RTT"
	acID [16]byte
	// Initialized to SEGGER_RTT_MAX_NUM_UP_BUFFERS (type. 2)
	MaxNumUpBuffers int32
	// Initialized to SEGGER_RTT_MAX_NUM_DOWN_BUFFERS (type. 2)
	MaxNumDownBuffers int32
	// Up buffers, transferring information up from target via debug probe to host
	aUp [MaxNumUpBuffers]BufferUpRTT
	// Down buffers, transferring information down from host via debug probe to target
	aDown [MaxNumDownBuffers]BufferDownRTT
	// stored slice's headers for gophers way to manipulate data...
	sliceUp       [MaxNumUpBuffers][]uint8
	sliceDown     [MaxNumDownBuffers][]uint8
	currentTermId byte
}

var (
	_RTT           ControlBlockRTT
	termName       = [...]byte{'T', 'e', 'r', 'm', 'i', 'n', 'a', 'l', '\x00'}
	termID         = [...]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F'}
	NotEnoughSpace = errors.New("not enough room")
)

func InitRtt(upSize, downSize uint32) {
	upBuffer := make([]uint8, upSize)
	downBuffer := make([]uint8, downSize)

	_RTT.MaxNumDownBuffers = MaxNumDownBuffers
	_RTT.MaxNumUpBuffers = MaxNumUpBuffers
	up := BufferUpRTT{
		sName:        &termName[0],
		pBuffer:      &upBuffer[0],
		sizeOfBuffer: upSize,
		flags:        0, // todo > add modes (SKIP, TRIM...)
	}
	down := BufferDownRTT{
		sName:        &termName[0],
		pBuffer:      &downBuffer[0],
		sizeOfBuffer: downSize,
		flags:        0, // todo > add modes (SKIP, TRIM...)
	}

	_RTT.aUp[0] = up
	_RTT.sliceUp[0] = upBuffer

	_RTT.aDown[0] = down
	_RTT.sliceDown[0] = downBuffer

	_RTT.currentTermId = 0
	copy(_RTT.acID[7:], []byte("RTT"))
	copy(_RTT.acID[0:], []byte("SEGGER"))
	_RTT.acID[6] = ' '
}

// Terminal is always channel 0, however it support virtual terminals from 0 to 15
type Terminal struct {
	id byte
}

func NewTerminal(id uint8) *Terminal {
	// RTT Viewer supports only 0..15 terminal id's for channel 0
	if id > 15 {
		return nil
	}
	// The RTT control block was not init yet.
	// Will init it with default size buffers (SizeUpBuffer and SizeDownBuffer).
	if _RTT.acID[0] == 0x00 {
		InitRtt(SizeUpBuffer, SizeDownBuffer)
	}

	return &Terminal{id: termID[id]}
}

func (t *Terminal) WriteString(s string) (int, error) {
	return t.Write([]byte(s))
}

func (t *Terminal) Write(s []byte) (int, error) {
	wrOff := _RTT.aUp[0].wrOff
	rdOFF := _RTT.aUp[0].rdOff

	var availSpace uint32

	// find available space
	switch {
	case rdOFF <= wrOff:
		availSpace = _RTT.aUp[0].sizeOfBuffer - (wrOff - rdOFF) - 1
	default:
		availSpace = (rdOFF - wrOff) - 1
	}

	len := uint32(len(s))

	if len == 0 {
		return 0, nil
	}

	if availSpace < len {
		return 0, NotEnoughSpace
	}

	// changing terminal if required
	if _RTT.currentTermId != t.id {

		// do we have space for changing terminal and msg?
		if availSpace < (len + 2) {
			return 0, NotEnoughSpace
		}

		_RTT.currentTermId = t.id
		var s = []byte{
			0xff,
			t.id,
		}
		t.Write(s)
		availSpace -= 2
		wrOff = _RTT.aUp[0].wrOff
	}

	if wrOff < rdOFF {
		copy(_RTT.sliceUp[0][wrOff:], s)
	} else {
		cntWrapAround := _RTT.aUp[0].sizeOfBuffer - wrOff
		copy(_RTT.sliceUp[0][wrOff:], s)
		if cntWrapAround < len {
			copy(_RTT.sliceUp[0][:(availSpace-cntWrapAround)], s[cntWrapAround:])
		}
	}

	_RTT.aUp[0].wrOff = (wrOff + len) % _RTT.aUp[0].sizeOfBuffer
	return int(len), nil
}
