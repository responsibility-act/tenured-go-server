package protocol

import (
	"encoding/json"
	"fmt"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/atomic"
)

const ACK = 010
const ONEWAY = 001

const NO_HEADER = commons.Error("NoHeader")
const SUCCESS = 0

var atomicId atomic.AtomicUInt32

func init() {
	atomicId = atomic.AtomicUInt32{}
}

type tenuredCommand struct {
	//消息ID，每个消息均要发送且每次发送都要递增，这样用户区分用户发送的消息。当发送消息达到最大后从零开始（如果可能的话）
	Id uint32

	//消息类型，查阅RequestCode，如果（flag & 0b10 > 0）&& code != 0，返回请求错误值ResponseCode。
	Code uint16

	//当前消息版本号，用户兼容相同消息的不同的版本(预留了255个可升级)
	Version uint8

	//第一位标识是否是请求，0=请求，1=ACK。第二位：0:不是，1：单向通知，不需要回复
	Flag int

	//	消息内容header，此处内容用户出消息头的传递。可以为空
	Header []byte

	//	消息内容body，用户传递消息内容体字节流，如果消息类型是ACK且code != 0 此处传递是错误消息内容描述，且不经过base64处理。可用为空
	Body []byte
}

func (this *tenuredCommand) String() string {
	return fmt.Sprintf("id=%d, code=%d, version=%d, flag=%d, header:%s, body:%v", this.Id, this.Code, this.Version, this.Flag, string(this.Header), this.Body)
}

func (this *tenuredCommand) IsSuccess() bool {
	return this.Code == SUCCESS
}
func (this *tenuredCommand) IsACK() bool {
	return (this.Flag & ACK) == ACK
}

func (this *tenuredCommand) IsOneway() bool {
	return (this.Flag & ONEWAY) == ONEWAY
}

func (this *tenuredCommand) MakeACK() *tenuredCommand {
	this.Flag = this.Flag | ACK
	return this
}

func (this *tenuredCommand) MakeOneway() *tenuredCommand {
	this.Flag = this.Flag | ONEWAY
	return this
}

func (this *tenuredCommand) SetHeader(header interface{}) error {
	if header == nil {
		return NO_HEADER
	}
	if bs, err := json.Marshal(header); err != nil {
		return err
	} else {
		this.Header = bs
		return nil
	}
}

func (this *tenuredCommand) GetHeader(header interface{}) error {
	if header == nil {
		return NO_HEADER
	}
	return json.Unmarshal(this.Header, header)
}

func (this *tenuredCommand) Error(error, message string) *tenuredCommand {
	this.Code = uint16(1)
	this.Header = []byte(error)
	this.Body = []byte(message)
	return this
}

func (this *tenuredCommand) RemotingError(error commons.RemotingError) *tenuredCommand {
	return this.Error(error.Code, error.Message)
}

func (this *tenuredCommand) GetError() *commons.RemotingError {
	if !this.IsACK() || this.IsSuccess() {
		return nil
	}
	return &commons.RemotingError{
		Code:    string(this.Header),
		Message: string(this.Body),
	}
}

func NewRequest(code uint16) *tenuredCommand {
	rc := &tenuredCommand{}
	rc.Id = atomicId.IncrementAndGet()
	rc.Code = code
	return rc
}

func NewACK(id uint32) *tenuredCommand {
	rc := &tenuredCommand{}
	rc.Id = id
	rc.Code = SUCCESS
	rc.MakeACK()
	return rc
}
