package introspect

import (
	"fmt"

	"github.com/xo/dburl"
)

// Database represents an introspected database.
type Database struct {
	Driver  string
	Schemas map[string]*Schema
}

// Schema represents an introspected database schema.
type Schema struct {
	Tables map[string]*Table
}

// Table represents an introspected database table.
type Table struct {
	Type    string
	Columns []*Column
}

// Column represents an introspected database table column.
type Column struct {
	Name string
	Type string
}

// Options represents the options for database introspection.
type Options struct {
	// Which schemas to introspect. If empty, the driver decide which schemas to introspect.
	Schemas []string
	// Which tables to ignore. If empty, the driver introspects all tables.
	IgnoredTables []string
	// Tables to include. If not empty, the driver only introspects the specified tables.
	Tables []string
}

// Introspector is an interface for database introspection. Each supported driver should implement this interface.
type Introspector interface {
	Introspect() (*Database, error)
}

// Introspect takes a database URL and returns a Database struct representing the
// introspected database, based on the options provided.
func Introspect(dbURL string, options *Options) (*Database, error) {
	// parse the URL first to get the driver name from the scheme
	url, err := dburl.Parse(dbURL)
	if err != nil {
		return nil, err
	}

	db, err := dburl.Open(dbURL)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	driverName := url.Scheme
	switch driverName {
	case "pg", "postgres", "pgsql":
		return NewPG(db, options).Introspect()
	// TODO(patrickdevivo): add support for other drivers
	default:
		return nil, fmt.Errorf("unsupported driver %s", driverName)
	}
}
