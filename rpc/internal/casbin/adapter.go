package casbin

import (
	"context"

	commonadapter "github.com/coder-lulu/newbee-common/casbin/adapter"
	"github.com/coder-lulu/newbee-core/rpc/ent"
)

// 🔥 使用common包的EntAdapter，保持向后兼容
// 通过类型别名和工厂函数，无需修改调用代码

// EntAdapter 类型别名，指向common包的实现
type EntAdapter = commonadapter.EntAdapter

// NewEntAdapter 工厂函数，创建适配器
// 🔥 内部使用EntCasbinRuleQuerier查询数据库
func NewEntAdapter(db *ent.Client, ctx context.Context) *EntAdapter {
	querier := NewEntCasbinRuleQuerier(db)
	return commonadapter.NewEntAdapter(querier, ctx)
}