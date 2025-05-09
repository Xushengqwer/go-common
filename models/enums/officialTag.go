package enums

// OfficialTag 官方标签枚举
// - 使用场景: 表示帖子的官方标签状态，用于展示帖子的可信度或特性
// - 枚举值:
//   - 0: 无标签 (默认值)
//   - 1: 官方认证
//   - 2: 预付保证金
//   - 3: 急速响应
type OfficialTag int

const (
	OfficialTagNone      OfficialTag = 0 // 无标签
	OfficialTagCertified OfficialTag = 1 // 官方认证
	OfficialTagDeposit   OfficialTag = 2 // 预付保证金
	OfficialTagRapid     OfficialTag = 3 // 急速响应
)
