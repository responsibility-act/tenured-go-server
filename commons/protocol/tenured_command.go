package protocol

import (
	"encoding/json"
	"fmt"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/atomic"
)

const FLAG_ACK = 2    //0b10
const FLAG_ONEWAY = 1 //0b01

const RESPONSE_SUCCESS = 0
const REQUEST_CODE_IDLE = uint16(0)
const REQUEST_CODE_ATUH = uint16(1)

const ErrNoHeader = commons.Error("NoHeader")

var atomicId atomic.AtomicUInt32

func init() {
	atomicId = atomic.AtomicUInt32(0)
}

type TenuredCommand struct {
	//消息ID，每个消息均要发送且每次发送都要递增，这样用户区分用户发送的消息。当发送消息达到最大后从零开始（如果可能的话）
	id uint32

	//消息类型，查阅RequestCode，如果（flag & 0b10 > 0）&& code != 0，返回请求错误值ResponseCode。
	code uint16

	//当前消息版本号，用户兼容相同消息的不同的版本(预留了255个可升级)
	Version uint8

	//第一位标识是否是请求，0=请求，1=FLAG_ACK。第二位：0:不是，1：单向通知，不需要回复
	flag int

	//	消息内容header，此处内容用户出消息头的传递。可以为空
	header []byte

	//	消息内容body，用户传递消息内容体字节流，如果消息类型是ACK且code != 0 此处传递是错误消息内容描述，且不经过base64处理。可用为空
	Body []byte
}

func (this *TenuredCommand) ID() uint32 {
	return this.id
}

func (this *TenuredCommand) String() string {
	return fmt.Sprintf("id=%d, code=%d, version=%d, flag=%d, header:%s, body:%v", this.id, this.code, this.Version, this.flag, string(this.header), this.Body)
}

func (this *TenuredCommand) IsSuccess() bool {
	return this.code == RESPONSE_SUCCESS
}
func (this *TenuredCommand) IsACK() bool {
	return (this.flag & FLAG_ACK) == FLAG_ACK
}

func (this *TenuredCommand) IsOneway() bool {
	return (this.flag & FLAG_ONEWAY) == FLAG_ONEWAY
}

func (this *TenuredCommand) MakeACK() *TenuredCommand {
	this.flag = this.flag | FLAG_ACK
	return this
}

func (this *TenuredCommand) MakeOneway() *TenuredCommand {
	this.flag = this.flag | FLAG_ONEWAY
	return this
}

func (this *TenuredCommand) SetHeader(header interface{}) error {
	if header == nil {
		return ErrNoHeader
	}
	if bs, err := json.Marshal(header); err != nil {
		return err
	} else {
		this.header = bs
		return nil
	}
}

func (this *TenuredCommand) GetHeader(header interface{}) error {
	if this.header == nil || header == nil {
		return ErrNoHeader
	}
	return json.Unmarshal(this.header, header)
}

func (this *TenuredCommand) Error(error, message string) *TenuredCommand {
	this.code = uint16(1)
	this.header = []byte(error)
	this.Body = []byte(message)
	return this
}

func (this *TenuredCommand) RemotingError(error *TenuredError) *TenuredCommand {
	if error != nil {
		return this.Error(error.Code, error.Message)
	} else {
		return this
	}
}

func (this *TenuredCommand) GetError() error {
	if !this.IsACK() || this.IsSuccess() {
		return nil
	}
	return &TenuredError{
		Code:    string(this.header),
		Message: string(this.Body),
	}
}

func NewRequest(code uint16) *TenuredCommand {
	rc := &TenuredCommand{}
	rc.id = atomicId.IncrementAndGet()
	rc.code = code
	return rc
}

func NewACK(id uint32) *TenuredCommand {
	rc := &TenuredCommand{}
	rc.id = id
	rc.code = RESPONSE_SUCCESS
	rc.MakeACK()
	return rc
}

func NewIdle() *TenuredCommand {
	return NewRequest(REQUEST_CODE_IDLE)
}
