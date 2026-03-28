package abstract

// Guard 守卫接口
type Guard interface {
	CanActivate(ctx Context) bool
}

// GuardFunc 守卫函数类型
type GuardFunc func(ctx Context) bool

// CanActivate 实现 Guard 接口
func (f GuardFunc) CanActivate(ctx Context) bool {
	return f(ctx)
}
