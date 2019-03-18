package protocol

import "time"

type TenuredClient struct {
}

func (this *TenuredClient) Invoke(channel string, command *TenuredCommand, timeout time.Duration) (*TenuredCommand, error) {
	return nil, nil
}

func (this *TenuredClient) AsyncInvoke(channel string, command *TenuredCommand, timeout time.Duration,
	callback func(tenuredCommand *TenuredCommand, err error)) {

}
