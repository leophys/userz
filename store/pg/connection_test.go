package pg

import (
	"context"
	sqllib "database/sql"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgproto3"
	"github.com/jackc/pgx/v4"

	"github.com/leophys/userz"
)

var _ interface {
	db
	transacter
} = &mockDB{}

var sqlRe *regexp.Regexp = regexp.MustCompile("\\$\\d+")

type mockDB struct {
	query       map[string]pgx.Rows
	queryRow    map[string]pgx.Row
	exec        map[string]string
	transacting bool
	commits     int
	rollback    int
}

func (db *mockDB) Prepare(ctx context.Context, name, statement string) (*pgconn.StatementDescription, error) {
	return nil, nil
}

func (db *mockDB) Begin(context.Context) (pgx.Tx, error) {
	db.transacting = true
	return db, nil
}

func (db *mockDB) Commit(context.Context) error {
	db.transacting = false
	db.commits++
	return nil
}

func (db *mockDB) Rollback(context.Context) error {
	if !db.transacting {
		return nil
	}
	db.transacting = false
	db.rollback++
	return nil
}

func (db *mockDB) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	statement := fmtSql(sql, args...)

	res, ok := db.exec[statement]
	if !ok {
		switch {
		case strings.HasPrefix(sql, "INSERT"):
			return pgconn.CommandTag([]byte("INSERT 0")), nil
		case strings.HasPrefix(sql, "UPDATE"):
			return pgconn.CommandTag([]byte("UPDATE 0")), nil
		case strings.HasPrefix(sql, "DELETE"):
			return pgconn.CommandTag([]byte("DELETE 0")), nil
		case strings.HasPrefix(sql, "SELECT"):
			return pgconn.CommandTag([]byte("SELECT 0")), nil
		default:
			return nil, fmt.Errorf("sql statement not understood: %s", sql)
		}
	}

	return pgconn.CommandTag([]byte(res)), nil
}

func (db *mockDB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	statement := fmtSql(sql, args...)

	rows, ok := db.query[statement]
	if !ok {
		return nil, fmt.Errorf("no rows for statement: %s", statement)
	}

	return rows, nil
}

func (db *mockDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	statement := fmtSql(sql, args...)

	row, ok := db.queryRow[statement]
	if !ok {
		return nil
	}

	return row
}

func (db *mockDB) BeginFunc(ctx context.Context, f func(pgx.Tx) error) (err error) { return nil }
func (db *mockDB) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (db *mockDB) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults { return nil }
func (db *mockDB) LargeObjects() (lo pgx.LargeObjects)                          { return }
func (db *mockDB) QueryFunc(ctx context.Context, sql string, args []interface{}, scans []interface{}, f func(pgx.QueryFuncRow) error) (pgconn.CommandTag, error) {
	return nil, nil
}
func (db *mockDB) Conn() *pgx.Conn { return nil }

func fmtSql(sql string, args ...interface{}) string {
	parts := strings.Split(
		sqlRe.ReplaceAllString(sql, "@"),
		"@",
	)

	if len(parts)-1 != len(args) {
		return ""
	}

	var result []string

	for i, arg := range args {
		t := reflect.TypeOf(arg)
		kind := t.Kind()

		switch kind {
		case reflect.Pointer, reflect.Array:
			if _, ok := arg.(interface {
				String() string
			}); ok {
				result = append(result, parts[i], "'%s'")
			}
		case reflect.Int, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
			result = append(result, parts[i], "%d")
		case reflect.String:
			result = append(result, parts[i], "'%s'")
		case reflect.Slice:
			if s, ok := arg.([]byte); ok {
				result = append(result, parts[i], "'%s'")
				args[i] = string(s)
			}
		case reflect.Struct:
			result = append(result, parts[i], "'%s'")
			switch t.Name() {
			case "Time":
				// noop (time.Time implements the Stringer interface)
			case "UUID":
				// noop (github.com/google/uuid.UUID implements the Stringer interface)
			case "NullString":
				args[i] = arg.(sqllib.NullString).String
			case "NullTime":
				args[i] = arg.(sqllib.NullTime).Time
			case "NullInt16":
				args[i] = arg.(sqllib.NullInt16).Int16
			case "NullInt32":
				args[i] = arg.(sqllib.NullInt32).Int32
			case "NullInt64":
				args[i] = arg.(sqllib.NullInt64).Int64
			}
		default:
			panic(fmt.Sprintf("unhandled type: %T", arg))
		}
	}
	result = append(result, parts[len(parts)-1])

	fmtStr := strings.Join(result, "")

	return fmt.Sprintf(fmtStr, args...)
}

type userRow userz.User

func (r *userRow) Scan(dest ...interface{}) error {
	if l := len(dest); l != 9 {
		return fmt.Errorf("wrong number of destination fields: %d", l)
	}

	id, err := uuid.Parse(r.Id)
	if err != nil {
		return err
	}

	*(dest[0].(*uuid.UUID)) = id
	*(dest[1].(*sqllib.NullString)) = sqllib.NullString{
		String: r.FirstName,
		Valid:  true,
	}
	*(dest[2].(*sqllib.NullString)) = sqllib.NullString{
		String: r.LastName,
		Valid:  true,
	}
	*(dest[3].(*string)) = r.NickName
	*(dest[4].(*[]byte)) = []byte(r.Password)
	*(dest[5].(*string)) = r.Email
	*(dest[6].(*sqllib.NullString)) = sqllib.NullString{
		String: r.Country,
		Valid:  true,
	}
	*(dest[7].(*sqllib.NullTime)) = sqllib.NullTime{
		Time:  r.CreatedAt,
		Valid: true,
	}
	*(dest[8].(*sqllib.NullTime)) = sqllib.NullTime{
		Time:  r.UpdatedAt,
		Valid: true,
	}

	return nil
}

type userRows struct {
	err  error
	cur  int
	rows []*userRow
}

func (rs *userRows) Close() {}

func (rs *userRows) Err() error {
	return rs.err
}

func (rs *userRows) CommandTag() pgconn.CommandTag {
	return nil
}

func (rs *userRows) FieldDescriptions() []pgproto3.FieldDescription { return nil }

func (rs *userRows) Next() bool {
	if rs.cur < len(rs.rows) {
		rs.cur++
		return true
	}
	return false
}

func (rs *userRows) Scan(dest ...interface{}) error {
	return rs.rows[rs.cur-1].Scan(dest...)
}

func (rs *userRows) Values() ([]interface{}, error) {
	return nil, nil
}

func (rs *userRows) RawValues() [][]byte {
	return nil
}

var _ userz.Passworder = dummyHasher

func dummyHasher(password string) (userz.Password, error) {
	return []byte(password), nil
}
