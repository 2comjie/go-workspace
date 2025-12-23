package inet

type IConn interface {
	GetConnId() uint32
	RemoteAddr() string
	Write(data []byte) error
	Close()
}

type IService interface {
	OnConnStart(conn IConn)
	OnConnRead(conn IConn, readData []byte)
	OnConnStop(conn IConn)
}
