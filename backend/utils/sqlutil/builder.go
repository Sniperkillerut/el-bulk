package sqlutil

import (
	"fmt"
	"strings"
)

// DynamicQueryBuilder simplifies the construction of dynamic WHERE clauses with Postgres-style placeholders ($1, $2, etc.).
type DynamicQueryBuilder struct {
	BaseQuery  string
	Conditions []string
	Args       []interface{}
}

// NewBuilder initializes a new DynamicQueryBuilder.
func NewBuilder(base string) *DynamicQueryBuilder {
	return &DynamicQueryBuilder{
		BaseQuery: base,
	}
}

// AddCondition appends a condition to the query and tracks the argument.
func (b *DynamicQueryBuilder) AddCondition(clause string, arg interface{}) {
	placeholder := fmt.Sprintf("$%d", len(b.Args)+1)
	condition := strings.ReplaceAll(clause, "?", placeholder)
	b.Conditions = append(b.Conditions, condition)
	b.Args = append(b.Args, arg)
}

// Build returns the final query string and arguments.
func (b *DynamicQueryBuilder) Build() (string, []interface{}) {
	query := b.BaseQuery
	if len(b.Conditions) > 0 {
		if strings.Contains(strings.ToUpper(query), "WHERE") {
			query += " AND "
		} else {
			query += " WHERE "
		}
		query += strings.Join(b.Conditions, " AND ")
	}
	return query, b.Args
}
