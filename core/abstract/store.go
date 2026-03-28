package abstract

// ValueStore 值存储接口
type ValueStore interface {
	Set(key string, value any)
	Get(key string) any
}
