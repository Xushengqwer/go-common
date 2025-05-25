package kafkaevents

import (
	"github.com/Xushengqwer/go-common/models/enums"
	"time"
)

// ==================================
// 1. 共享核心数据结构
// ==================================

// RejectionDetail 封装了审核拒绝的具体原因。
type RejectionDetail struct {
	Label          string   `json:"label"`
	Suggestion     string   `json:"suggestion,omitempty"`
	Score          float64  `json:"score,omitempty"`
	MatchedContent []string `json:"matchedContent,omitempty"`
}

// PostData 是跨 Kafka 事件共享的、统一的帖子数据核心结构。
type PostData struct {
	ID             uint64            `json:"id"`
	Title          string            `json:"title"`
	Content        string            `json:"content"`
	AuthorID       string            `json:"author_id"`
	AuthorAvatar   string            `json:"author_avatar"`
	AuthorUsername string            `json:"author_username"`
	Tags           []string          `json:"tags,omitempty"`
	Status         enums.Status      `json:"status"` // <-- 使用 enums!
	ViewCount      int64             `json:"view_count"`
	OfficialTag    enums.OfficialTag `json:"official_tag"` // <-- 使用 enums!
	PricePerUnit   float64           `json:"price_per_unit"`
	ContactQRCode  string            `json:"contact_qr_code,omitempty"`
	CreatedAt      int64             `json:"created_at"`
	UpdatedAt      int64             `json:"updated_at"`
}

// ==================================
// 2. 统一的 Kafka 事件结构
// ==================================

// --- 事件元数据 (可选但推荐) ---
// 可以考虑为每个事件添加统一的元数据字段，例如 EventID 和 Timestamp

// PostPendingAuditEvent 当帖子创建或更新需要审核时，由 post-service 发布。
type PostPendingAuditEvent struct {
	EventID   string    `json:"event_id"`  // 事件唯一ID
	Timestamp time.Time `json:"timestamp"` // 事件发生时间
	Post      PostData  `json:"post"`      // 完整的帖子数据
}

// PostApprovedEvent 当帖子审核通过时，由 audit-service 发布。
type PostApprovedEvent struct {
	EventID   string    `json:"event_id"`  // 事件唯一ID
	Timestamp time.Time `json:"timestamp"` // 事件发生时间
	Post      PostData  `json:"post"`      // 完整的帖子数据 (供 ES 和 post-service 使用)
}

// PostRejectedEvent 当帖子审核不通过时，由 audit-service 发布。
type PostRejectedEvent struct {
	EventID    string            `json:"event_id"`  // 事件唯一ID
	Timestamp  time.Time         `json:"timestamp"` // 事件发生时间
	PostID     uint64            `json:"post_id"`   // 帖子ID
	Suggestion string            `json:"suggestion"`
	Details    []RejectionDetail `json:"details"`
}

// PostDeletedEvent 当帖子被删除时，由 post-service 发布。
type PostDeletedEvent struct {
	EventID   string    `json:"event_id"`  // 事件唯一ID
	Timestamp time.Time `json:"timestamp"` // 事件发生时间
	PostID    uint64    `json:"post_id"`   // 被删除的帖子ID
}

// DeadLetterEvent (可选，如果希望在 common 中统一 DLQ 结构)
// type DeadLetterEvent struct { ... }
