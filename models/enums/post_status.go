package enums

// Stats 定义状态枚举类型
type Stats uint

// 定义枚举常量
const (
	Pending  Stats = iota // 0 待审核
	Approved              // 1 审核通过
	Rejected              // 2 拒绝
)
