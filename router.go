package jinx

type Router interface {
	// BeforeHandle 处理业务前的方法
	BeforeHandle(request Request)
	// Handle 具体执行的方法
	Handle(request Request)
	// AfterHandle 处理业务后的方法
	AfterHandle(request Request)
}

type BaseRouter struct {
}

func (b *BaseRouter) BeforeHandle(request Request) {
}

func (b *BaseRouter) Handle(request Request) {
}

func (b *BaseRouter) AfterHandle(request Request) {
}
