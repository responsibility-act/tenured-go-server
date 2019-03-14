package atomic

import "sync/atomic"

type AtomicUInt32 uint32

func (self *AtomicUInt32) Get() uint32 {
	return atomic.LoadUint32((*uint32)(self))
}

func (self *AtomicUInt32) IncrementAndGet() uint32 {
	return self.AddAndGet(1)
}

func (self *AtomicUInt32) GetAndIncrement() uint32 {
	return self.GetAndAdd(1)
}

func (self *AtomicUInt32) DecrementAndGet() uint32 {
	return self.AddAndGet(-1)
}

func (self *AtomicUInt32) GetAndDecrement() uint32 {
	return self.GetAndAdd(-1)
}

func (self *AtomicUInt32) AddAndGet(i int) uint32 {
	var ret uint32
	for {
		ret = atomic.LoadUint32((*uint32)(self))
		if atomic.CompareAndSwapUint32((*uint32)(self), ret, ret+uint32(i)) {
			break
		}
	}
	return ret + 1
}

func (self *AtomicUInt32) GetAndAdd(i int) uint32 {
	var ret uint32
	for {
		ret = atomic.LoadUint32((*uint32)(self))
		if atomic.CompareAndSwapUint32((*uint32)(self), ret, ret+uint32(i)) {
			break
		}
	}
	return ret
}

func (self *AtomicUInt32) Set(i int) {
	atomic.StoreUint32((*uint32)(self), uint32(i))
}

func (self *AtomicUInt32) CompareAndSet(expect int, update int) bool {
	return atomic.CompareAndSwapUint32((*uint32)(self), uint32(expect), uint32(update))
}

func NewUint32(initValue uint32) *AtomicUInt32 {
	u := AtomicUInt32(initValue)
	return &u
}
