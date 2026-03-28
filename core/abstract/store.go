package abstract

// ValueStoreAbstract 值存储接口
type ValueStoreAbstract interface {
	Set(key string, value any)
	Get(key string) any
}
