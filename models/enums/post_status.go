package enums

// Status 状态枚举
// - 使用场景: 表示各种需要审核或有明确状态流转的对象的当前状态，例如帖子审核、用户审核等。
// - 枚举值:
//   - 0: 待审核 (Pending) - 对象已提交，等待处理。
//   - 1: 审核通过 (Approved) - 对象已通过审核，变为有效或激活状态。
//   - 2: 拒绝 (Rejected) - 对象未通过审核，变为无效或非激活状态。
type Status int

// 定义枚举常量
const (
	Pending  Status = 0 // 0 待审核 - 对象已提交，等待处理。
	Approved Status = 1 // 1 审核通过 - 对象已通过审核，变为有效或激活状态。
	Rejected Status = 2 // 2 拒绝 - 对象未通过审核，变为无效或非激活状态。
)
