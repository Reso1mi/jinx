package jinx

type Handler interface {
	OnOpen(f func(c *connection))

	OnClose(f func(c *connection))

	OnRead(f func(c *connection))

	OnWrite(f func(c *connection))

	OnData(f func(c *connection))
}

type handler struct {
	onOpen func(c *connection)

	onClose func(c *connection)

	onRead func(c *connection)

	onWrite func(c *connection)

	onData func(c *connection)
}
