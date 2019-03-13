package protocol

import (
	"encoding/binary"
	"errors"
	"github.com/ihaiker/tenured-go-server/commons/remoting"
	"io"
	"os"
	"strconv"
)

const LENGTH_MIN = 4 /*length*/ + 4 /*id*/ + 2 /*code*/ + 1 /*version*/ + 3 /*(header.length < 2) | flag*/
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
	}
	//id
	if err := binary.Read(reader, endian, &command.Id); err != nil {
		return nil, err
	}
	//code
	if err := binary.Read(reader, endian, &command.Code); err != nil {
		return nil, err
	}
	vf := uint32(0)
	if err := binary.Read(reader, endian, &vf); err != nil {
		return nil, err
	}
	command.Version = uint8((vf >> 24) & 0xFF)
	command.Flag = int(vf & 011)
	headerLength := int((vf & 0xfff) >> 2)
	if headerLength > 0 {
		command.Header = make([]byte, headerLength)
		if i, err := reader.Read(command.Header); err != nil {
			return nil, &remoting.RemotingError{Op: remoting.ErrDecoder, Err: err}
		} else if i != headerLength {
			return nil, &remoting.RemotingError{Op: remoting.ErrDecoder, Err: errors.New("head length")}
		}
	}
	bodyLength := int(length) - LENGTH_MIN - headerLength
	if bodyLength > 0 {
		command.Body = make([]byte, bodyLength)
		if i, err := reader.Read(command.Body); err != nil {
			return nil, &remoting.RemotingError{Op: remoting.ErrDecoder, Err: err}
		} else if i != bodyLength {
			return nil, &remoting.RemotingError{Op: remoting.ErrDecoder, Err: errors.New("body length")}
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
	length := uint32(LENGTH_MIN)
	headerLength := uint32(0)

	if msg.Header != nil && len(msg.Header) != 0 {
		headerLength = uint32(len(msg.Header))
	}
	length += headerLength

	if msg.Body != nil && len(msg.Body) != 0 {
		length += uint32(len(msg.Body))
	}
	if int64(length) > int64(this.config.PacketBytesLimit) {
		return nil, &remoting.RemotingError{Op: remoting.ErrPacketBytesLimit,
			Err: errors.New("the packet limit size " + strconv.Itoa(this.config.PacketBytesLimit))}
	}

	bs := make([]byte, length)
	endian.PutUint32(bs, length)       //4
	endian.PutUint32(bs[4:], msg.Id)   //4
	endian.PutUint16(bs[8:], msg.Code) //2

	vf := headerLength
	vf = (uint32(msg.Version&0xFF) << 24) | (vf << 2) | uint32(msg.Flag)
	endian.PutUint32(bs[10:], vf)

	if headerLength > 0 {
		copy(bs[14:], msg.Header)
	}
	if msg.Body != nil && len(msg.Body) > 0 {
		copy(bs[14+headerLength:], msg.Body)
	}
	return bs, nil
}
