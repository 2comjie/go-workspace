package codec

import (
	"encoding/binary"
	"hutool/bitx"
)

const (
	HeartbeatBitPos = 1
	IsOneWayBitPos  = 2
)

const (
	IsPushPacketBitPost = 1
)

type S2CPacket struct {
	bytes []byte
}

func NewS2CRspPacket(reqId uint32, body []byte) S2CPacket {
	var head byte = 0
	bitx.SetBit(&head, IsPushPacketBitPost, false)
	bytes := make([]byte, len(body)+5)
	bytes[0] = head

	binary.BigEndian.PutUint32(bytes[1:5], reqId)
	copy(bytes[5:], body)
	return S2CPacket{bytes: bytes}
}

func NewS2CPushPacket(serviceId uint32, routerId uint32, body []byte) S2CPacket {
	var head byte = 0
	bitx.SetBit(&head, IsPushPacketBitPost, true)
	bytes := make([]byte, len(body)+9)
	bytes[0] = head

	binary.BigEndian.PutUint32(bytes[1:5], serviceId)
	binary.BigEndian.PutUint32(bytes[5:9], routerId)
	copy(bytes[9:], body)
	return S2CPacket{bytes: bytes}
}

func BytesToS2CPacket(bytes []byte) (S2CPacket, error) {
	p := S2CPacket{bytes: nil}
	if len(bytes) < 1 {
		return p, PacketBytesErr
	}
	head := bytes[0]
	if bitx.IsBitSet(head, IsPushPacketBitPost) {
		if len(bytes) < 9 {
			return p, PacketBytesErr
		}
	} else {
		if len(bytes) < 5 {
			return p, PacketBytesErr
		}
	}

	p.bytes = bytes
	return p, nil
}

func (p S2CPacket) IsPushPacket() bool {
	return bitx.IsBitSet(p.bytes[0], IsPushPacketBitPost)
}

func (p S2CPacket) ServiceId() uint32 {
	return binary.BigEndian.Uint32(p.bytes[1:5])
}

func (p S2CPacket) RouterId() uint32 {
	return binary.BigEndian.Uint32(p.bytes[5:9])
}

func (p S2CPacket) ReqId() uint32 {
	return binary.BigEndian.Uint32(p.bytes[1:5])
}

func (p S2CPacket) Body() []byte {
	if p.IsPushPacket() {
		return p.bytes[9:]
	}
	return p.bytes[5:]
}

func (p S2CPacket) Bytes() []byte {
	return p.bytes
}

type C2SPacket struct {
	bytes []byte
}

func NewC2SHeartBeatPacket() *C2SPacket {
	var head byte = 0
	bitx.SetBit(&head, HeartbeatBitPos, true)
	bitx.SetBit(&head, IsOneWayBitPos, true)
	return &C2SPacket{bytes: []byte{head}}
}

func NewC2SReqPacket(serviceId uint32, routerId uint32, reqId uint32, isOneWay bool, body []byte) C2SPacket {
	var head byte = 0
	bitx.SetBit(&head, IsOneWayBitPos, isOneWay)
	bitx.SetBit(&head, HeartbeatBitPos, false)
	bytes := make([]byte, 13)
	bytes[0] = head
	binary.BigEndian.PutUint32(bytes[1:5], serviceId)
	binary.BigEndian.PutUint32(bytes[5:9], routerId)
	binary.BigEndian.PutUint32(bytes[9:13], reqId)
	bytes = append(bytes, body...)
	return C2SPacket{bytes: bytes}
}

func BytesToC2SPacket(bytes []byte) (C2SPacket, error) {
	p := C2SPacket{bytes: nil}
	if len(bytes) < 1 {
		return p, PacketBytesErr
	}
	head := bytes[0]
	if bitx.IsBitSet(head, HeartbeatBitPos) {
		if len(bytes) != 1 {
			return p, PacketBytesErr
		}
	} else {
		if len(bytes) < 13 {
			return p, PacketBytesErr
		}
	}

	p.bytes = bytes
	return p, nil
}

func (p C2SPacket) IsHeartbeatPacket() bool {
	return bitx.IsBitSet(p.bytes[0], HeartbeatBitPos)
}

func (p C2SPacket) IsOneWay() bool {
	return bitx.IsBitSet(p.bytes[0], IsOneWayBitPos)
}

func (p C2SPacket) ServiceId() uint32 {
	return binary.BigEndian.Uint32(p.bytes[1:5])
}

func (p C2SPacket) RouterId() uint32 {
	return binary.BigEndian.Uint32(p.bytes[5:9])
}

func (p C2SPacket) ReqId() uint32 {
	return binary.BigEndian.Uint32(p.bytes[9:13])
}

func (p C2SPacket) Body() []byte {
	return p.bytes[13:]
}

func (p C2SPacket) Bytes() []byte {
	return p.bytes
}
