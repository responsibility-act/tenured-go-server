package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/ihaiker/tenured-go-server/commons/remoting"
	"io"
	"os"
	"strconv"
)

const lengthMin = 4 /*length*/ + 4 /*id*/ + 2 /*code*/ + 1 /*version*/ + 3 /*(header.length < 2) | flag*/
var endian = binary.BigEndian

type tenuredCoder struct {
	config *remoting.RemotingConfig
}

func (this *tenuredCoder) Decode(channel remoting.RemotingChannel, reader io.Reader) (interface{}, error) {
	command := &TenuredCommand{}
	length := uint32(0)
	//length
	if err := binary.Read(reader, endian, &length); err != nil {
		return nil, err
	} else if length < uint32(lengthMin) || length >= uint32(this.config.PacketBytesLimit) {
		return nil, &remoting.RemotingError{Op: remoting.ErrDecoder, Err: errors.New(fmt.Sprintf("head length %d", length))}
	}
	//id
	if err := binary.Read(reader, endian, &command.id); err != nil {
		return nil, err
	}
	//code
	if err := binary.Read(reader, endian, &command.code); err != nil {
		return nil, err
	}
	vf := uint32(0)
	if err := binary.Read(reader, endian, &vf); err != nil {
		return nil, err
	}

	command.Version = uint8((vf >> 24) & 0xFF)
	command.flag = int(vf & 3 /*0b11*/)
	headerLength := int((vf & 0x3FFFFF) >> 2)
	if headerLength > 0 {
		command.header = make([]byte, headerLength, headerLength)
		if i, err := reader.Read(command.header); err != nil {
			return nil, &remoting.RemotingError{Op: remoting.ErrDecoder, Err: err}
		} else if i != headerLength {
			return nil, &remoting.RemotingError{Op: remoting.ErrDecoder,
				Err: errors.New(fmt.Sprintf("head length export %d bug %d", headerLength, i))}
		}
	}
	bodyLength := int(length) - lengthMin - headerLength
	if bodyLength > 0 {
		command.Body = make([]byte, bodyLength)
		if i, err := reader.Read(command.Body); err != nil {
			return nil, &remoting.RemotingError{Op: remoting.ErrDecoder, Err: err}
		} else if i != bodyLength {
			return nil, &remoting.RemotingError{Op: remoting.ErrDecoder,
				Err: errors.New(fmt.Sprintf("body length export %d bus %d", bodyLength, i))}
		}
	}
	return command, nil
}

func (this *tenuredCoder) Encode(channel remoting.RemotingChannel, msg interface{}) ([]byte, error) {
	if bs, ok := msg.(*TenuredCommand); ok {
		return this.encodeCommand(bs)
	} else {
		return nil, os.ErrInvalid
	}
}

func (this *tenuredCoder) encodeCommand(msg *TenuredCommand) ([]byte, error) {
	length := uint32(lengthMin)
	headerLength := uint32(0)

	if msg.header != nil && len(msg.header) != 0 {
		headerLength = uint32(len(msg.header))
	}
	length += headerLength

	if msg.Body != nil && len(msg.Body) != 0 {
		length += uint32(len(msg.Body))
	}
	if int64(length) > int64(this.config.PacketBytesLimit) {
		return nil, &remoting.RemotingError{Op: remoting.ErrPacketBytesLimit,
			Err: errors.New("the packet limit size " + strconv.Itoa(this.config.PacketBytesLimit))}
	}

	bs := make([]byte, length, length)
	endian.PutUint32(bs, length)       //4
	endian.PutUint32(bs[4:], msg.id)   //4
	endian.PutUint16(bs[8:], msg.code) //2

	vf := (uint32(msg.Version&0xFF) << 24) | uint32((headerLength&0x3FFFFF)<<2) | uint32(msg.flag&3 /*0b11*/)
	endian.PutUint32(bs[10:], vf)

	if headerLength > 0 {
		copy(bs[14:], msg.header)
	}
	if msg.Body != nil && len(msg.Body) > 0 {
		copy(bs[14+headerLength:], msg.Body)
	}
	return bs, nil
}
