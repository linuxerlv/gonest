package abstract

// GuardAbstract 守卫接口
type GuardAbstract interface {
	CanActivate(ctx ContextAbstract) bool
}

// GuardFuncAbstract 守卫函数类型
type GuardFuncAbstract func(ctx ContextAbstract) bool

// CanActivate 实现 GuardAbstract 接口
func (f GuardFuncAbstract) CanActivate(ctx ContextAbstract) bool {
	return f(ctx)
}
