package abstract

// Env 环境变量接口
// 用于读取和管理系统环境变量，支持依赖注入和测试替换
type Env interface {
	// Get 获取环境变量，不存在返回空字符串
	Get(key string) string

	// GetOrDefault 获取环境变量，不存在返回默认值
	GetOrDefault(key, defaultValue string) string

	// Has 检查环境变量是否存在
	Has(key string) bool

	// All 返回所有环境变量
	All() map[string]string

	// Set 设置环境变量（主要用于测试场景）
	Set(key, value string)

	// Unset 删除环境变量（主要用于测试场景）
	Unset(key string)
}
