package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/coder-lulu/newbee-common/orm/ent/mixins"
	commonMixins "github.com/coder-lulu/newbee-common/orm/ent/mixins"
)

type DictionaryDetail struct {
	ent.Schema
}

func (DictionaryDetail) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").
			Comment("The title shown in the ui | 展示名称 （建议配合i18n）"),
		field.String("value").
			Comment("value | 值"),
		field.String("list_class").Comment("listClass | 列表样式").Optional().Default("default"),
		field.String("css_class").Comment("css_class | 样式").Optional().Default(""),
		field.Uint32("is_default").
			Comment("is_default | 是否为默认值").Optional().Default(0),
		field.Uint64("dictionary_id").Optional().
			Comment("Dictionary ID | 字典ID"),
	}
}

func (DictionaryDetail) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDMixin{},
		mixins.StatusMixin{},
		mixins.SortMixin{},
		commonMixins.TenantMixin{},
	}
}

func (DictionaryDetail) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("dictionaries", Dictionary.Type).Field("dictionary_id").Ref("dictionary_details").Unique(),
	}
}

func (DictionaryDetail) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("Dictionary Key/Value Table | 字典键值表"),
		entsql.Annotation{Table: "sys_dictionary_details"},
	}
}
