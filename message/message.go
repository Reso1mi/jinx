package message

// Message 属性
type Message struct {
	id   int
	len  int
	data []byte
}

func (m Message) GetID() int {
	panic("implement me")
}

func (m Message) GetData() []byte {
	panic("implement me")
}

func (m Message) GetLen() int {
	panic("implement me")
}
