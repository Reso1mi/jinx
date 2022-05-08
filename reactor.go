package jinx

type Reactor interface {
	Callback() error
	Close() error
}
