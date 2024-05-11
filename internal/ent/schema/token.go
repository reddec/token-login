package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/reddec/token-login/internal/types"
)

// Token holds the schema definition for the Token entity.
type Token struct {
	ent.Schema
}

// Fields of the Token.
func (Token) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.String("key_id").GoType(&types.KeyID{}).ValueScanner(field.TextValueScanner[*types.KeyID]{}).Unique(),
		field.Bytes("hash").NotEmpty(),
		field.String("user"),
		field.String("label").Default(""),
		field.String("path").Default("/**"),
		field.String("host"),
		field.JSON("headers", types.Headers{}).Optional(),
		field.Int64("requests").Default(0),
		field.Time("last_access_at").Default(time.Now),
	}
}

// Edges of the Token.
func (Token) Edges() []ent.Edge {
	return nil
}

func (Token) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user").StorageKey("user_idx_ent"),
		index.Fields("key_id").StorageKey("key_id_ent").Unique(),
	}
}

func (Token) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "token"},
	}
}
