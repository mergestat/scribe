package introspect

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type pgIntrospector struct {
	db      *sql.DB
	options *Options
}

var tableQuery = `SELECT
t.table_schema AS schema,
t.table_name,
t.table_type,
c.column_name,
c.ordinal_position,
c.is_nullable,
c.data_type,
c.udt_name,
pg_catalog.col_description(format('%s.%s', c.table_schema, c.table_name)::regclass::oid, c.ordinal_position) AS column_description
FROM information_schema.tables AS t
INNER JOIN information_schema.columns AS c ON (t.table_name = c.table_name AND t.table_schema = c.table_schema)
WHERE t.table_schema IN ('public')
`

type pgIntrospectionResultRow struct {
	TableSchema       string  `db:"schema"`
	TableName         string  `db:"table_name"`
	TableType         string  `db:"table_type"`
	ColumnName        string  `db:"column_name"`
	OrdinalPosition   int     `db:"ordinal_position"`
	IsNullable        string  `db:"is_nullable"`
	DataType          string  `db:"data_type"`
	UdtName           string  `db:"udt_name"`
	ColumnDescription *string `db:"column_description"`
}

// NewPG returns a new Introspector for PostgreSQL.
func NewPG(db *sql.DB, options *Options) Introspector {
	return &pgIntrospector{db: db, options: options}
}

// Introspect introspects the database and returns a Database struct representing the
// introspected database.
func (i *pgIntrospector) Introspect() (*Database, error) {
	rows, err := i.db.Query(tableQuery)
	if err != nil {
		return nil, fmt.Errorf("error introspecting database: %v", err)
	}

	introspected := &Database{
		Driver:  "pg",
		Schemas: make(map[string]*Schema),
	}

	for rows.Next() {
		var row pgIntrospectionResultRow
		if err := rows.Scan(&row.TableSchema, &row.TableName, &row.TableType, &row.ColumnName, &row.OrdinalPosition, &row.IsNullable, &row.DataType, &row.UdtName, &row.ColumnDescription); err != nil {
			return nil, fmt.Errorf("error scanning row during introspection: %v", err)
		}

		// if this is a new schema, add it to the database
		if _, ok := introspected.Schemas[row.TableSchema]; !ok {
			introspected.Schemas[row.TableSchema] = &Schema{
				Tables: make(map[string]*Table),
			}
		}

		// if this is a new table, add it to the schema
		if _, ok := introspected.Schemas[row.TableSchema].Tables[row.TableName]; !ok {
			introspected.Schemas[row.TableSchema].Tables[row.TableName] = &Table{
				Type: row.TableType,
			}
		}

		// add the column to the table
		cols := &introspected.Schemas[row.TableSchema].Tables[row.TableName].Columns
		*cols = append(*cols, &Column{
			Name: row.ColumnName,
			Type: row.DataType,
		})
	}

	return introspected, nil
}
