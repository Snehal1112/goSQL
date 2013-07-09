package translators

import (
	"github.com/quintans/goSQL/db"
	tk "github.com/quintans/toolkit"
	coll "github.com/quintans/toolkit/collection"

	"strconv"
	"strings"
)

type PostgreSQLTranslator struct {
	*GenericTranslator
}

func NewPostgreSQLTranslator() db.Translator {
	this := new(PostgreSQLTranslator)
	this.GenericTranslator = new(GenericTranslator)
	this.Init(this)
	this.QueryProcessorFactory = func() QueryProcessor { return NewQueryBuilder(this) }
	this.InsertProcessorFactory = func() InsertProcessor { return NewInsertBuilder(this) }
	this.UpdateProcessorFactory = func() UpdateProcessor { return NewPgUpdateBuilder(this) }
	this.DeleteProcessorFactory = func() DeleteProcessor { return NewDeleteBuilder(this) }
	return this
}

var _ db.Translator = &PostgreSQLTranslator{}

func (this *PostgreSQLTranslator) GetAutoKeyStrategy() db.AutoKeyStrategy {
	return db.AUTOKEY_RETURNING
}

func (this *PostgreSQLTranslator) GetPlaceholder(index int, name string) string {
	return "$" + strconv.Itoa(index+1)
}

// INSERT
func (this *PostgreSQLTranslator) GetSqlForInsert(insert *db.Insert) string {
	// insert generated by super
	sql := this.GenericTranslator.GetSqlForInsert(insert)

	// only ONE numeric id is allowed
	// if no value was defined for the key, it is assumed an auto number,
	// otherwise is a guid (or something else)
	if !insert.HasKeyValue {
		str := tk.NewStrBuffer()
		str.Add(sql, " RETURNING ", this.overrider.ColumnName(insert.GetTable().GetSingleKeyColumn()))
		sql = str.String()
	}

	return sql
}

func (this *PostgreSQLTranslator) TableName(table *db.Table) string {
	return strings.ToLower(table.GetName())
}

func (this *PostgreSQLTranslator) ColumnName(column *db.Column) string {
	return strings.ToLower(column.GetName())
}

//// UPDATE

type PgUpdateBuilder struct {
	UpdateBuilder
}

func NewPgUpdateBuilder(translator db.Translator) *PgUpdateBuilder {
	this := new(PgUpdateBuilder)
	this.Super(translator)
	return this
}

func (this *PgUpdateBuilder) Column(values coll.Map, tableAlias string) {
	for it := values.Iterator(); it.HasNext(); {
		entry := it.Next()
		column := entry.Key.(*db.Column)
		// use only not virtual columns
		if !column.IsVirtual() {
			token := entry.Value.(db.Tokener)
			this.columnPart.AddAsOne(
				this.translator.ColumnName(column),
				" = ", this.translator.Translate(db.UPDATE, token))
		}
	}
}

// TODO: implement PaginateSQL
func (this *PostgreSQLTranslator) PaginateSQL(query *db.Query, sql string) string {
	sb := tk.NewStrBuffer()
	if query.GetLimit() > 0 {
		sb.Add(sql, " LIMIT :", db.LIMIT_PARAM)
		query.SetParameter(db.LIMIT_PARAM, query.GetLimit())
		if query.GetSkip() > 0 {
			sb.Add(" OFFSET :", db.OFFSET_PARAM)
			query.SetParameter(db.OFFSET_PARAM, query.GetSkip())
		}
		return sb.String()
	}

	return sql
}
