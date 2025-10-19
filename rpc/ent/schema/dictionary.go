package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/coder-lulu/newbee-common/orm/ent/mixins"
	commonMixins "github.com/coder-lulu/newbee-common/orm/ent/mixins"
)

type Dictionary struct {
	ent.Schema
}

func (Dictionary) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").
			Comment("The title shown in the ui | 展示名称 （建议配合i18n）"),
		field.String("name").
			Comment("The name of dictionary for search | 字典搜索名称"),
		field.String("desc").
			Comment("The description of dictionary | 字典的描述").
			Optional(),
	}
}

func (Dictionary) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.StatusMixin{},
		commonMixins.TenantMixin{},
	}
}

func (Dictionary) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("dictionary_details", DictionaryDetail.Type),
	}
}

func (Dictionary) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "name").
			Unique(),
	}
}

func (Dictionary) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("Dictionary Table | 字典信息表"),
		entsql.Annotation{Table: "sys_dictionaries"},
	}
}
