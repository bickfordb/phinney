package phinney

// Helpers for database/sql for selecting/inserting/updating structs
import _ "github.com/lib/pq"
import "database/sql"
import "fmt"
import "os"
import "reflect"
import "regexp"
import "strconv"
import "strings"
import "time"

var DBEngine string = os.Getenv("DB_ENGINE")
var DBURL string = os.Getenv("DB_URL")

func DBConn() (db *sql.DB, err error) {
	db, err = sql.Open(DBEngine, DBURL)
	if db != nil {
		db.SetMaxIdleConns(5)
	}
	return
}

type root struct {
	db        *sql.DB
	relation  string
	structPtr interface{}
}

type SelectQuery struct {
	parent *SelectQuery
	root   *root
	clause *clause
}

type clause struct {
	op    string
	key   string
	value interface{}
}

// result, err = Select("users", &user).Filter("Id", 5).Filter("Name", "Brandon").Exec()
// err = query.Exec()
// for query.Next() {
// }
// query.One()

func Select(db *sql.DB, relation string, structPtr interface{}) (query *SelectQuery) {
	return &SelectQuery{
		root: &root{
			db:        db,
			relation:  relation,
			structPtr: structPtr}}
}

func (q *SelectQuery) Filter(key string, value interface{}) (result *SelectQuery) {
	result = &SelectQuery{
		parent: q,
		clause: &clause{
			op:    "=",
			key:   key,
			value: value}}
	return
}

func (q *SelectQuery) whereSQL() (sql string, bind []interface{}) {
	for curr := q; curr != nil && curr.clause != nil; curr = curr.parent {
		if sql != "" {
			sql += " AND "
		}
		bind = append(bind, curr.clause.value)
		sql += fmt.Sprintf(`"%s" %s $%d`, capWordsToSnakeCase(curr.clause.key), curr.clause.op, len(bind))
	}
	return
}

func (q *SelectQuery) getRoot() (result *root) {
	for curr := q; curr != nil; curr = curr.parent {
		if curr.root != nil {
			result = curr.root
			break
		}
	}
	return
}

func (q *SelectQuery) fields() (result []*reflect.StructField) {
	t := reflect.ValueOf(q.getRoot().structPtr).Elem().Type()
	for i := 0; i < t.NumField(); i += 1 {
		f := t.Field(i)
		result = append(result, &f)
	}
	return
}

func (q *SelectQuery) keys() (result []string) {
	for _, f := range q.fields() {
		result = append(result, f.Name)
	}
	return
}

func (q *SelectQuery) colSQL() (sql string) {
	for i, key := range q.keys() {
		if i > 0 {
			sql += ", "
		}
		sql += fmt.Sprintf(`"%s"`, capWordsToSnakeCase(key))
	}
	return
}
func (q *SelectQuery) fromSQL() (result string) {
	return fmt.Sprintf(`"%s"`, q.getRoot().relation)

}

func (q *SelectQuery) Compile() (sql string, bind []interface{}) {
	sql = "SELECT "
	sql += q.colSQL()
	sql += " FROM "
	sql += q.fromSQL()
	whereSQL, bind := q.whereSQL()
	if whereSQL != "" {
		sql += " WHERE "
		sql += whereSQL
	}
	return
}

func (q *SelectQuery) exec(db *sql.DB) (result *SelectResult, err error) {
	sql, bind := q.Compile()
	log.Debug("sql: %q", sql)
	rows, err := db.Query(sql, bind...)
	if err != nil {
		return
	}
	result = &SelectResult{rows: rows, query: q}
	return
}

func (q *SelectQuery) All() (rows interface{}, err error) {
	db := q.getRoot().db
	if db == nil {
		db, err = DBConn()
		if err != nil {
			return
		}
		defer db.Close()
	}
	r, err := q.exec(db)
	if err != nil {
		return
	}
	defer r.Close()
  valuePtr := reflect.ValueOf(q.getRoot().structPtr)
  aSlice := reflect.MakeSlice(reflect.SliceOf(valuePtr.Elem().Type()), 0, 0)
  for {
    var isNext bool
    isNext, err = r.Next()
    if err != nil {
      return
    }
    if !isNext {
      break
    }
    aSlice = reflect.Append(aSlice, valuePtr.Elem())
  }
  rows = aSlice.Interface()
	return
}

func (q *SelectQuery) One() (isNext bool, err error) {
	db := q.getRoot().db
	if db == nil {
		db, err = DBConn()
		if err != nil {
			return
		}
		defer db.Close()
	}
	r, err := q.exec(db)
	if err != nil {
		return
	}
	defer r.Close()
	isNext, err = r.Next()
	return
}

type SelectResult struct {
	rows  *sql.Rows
	query *SelectQuery
}

func (s *SelectResult) Close() (err error) {
	return s.rows.Close()
}

var timeType reflect.Type = reflect.TypeOf(time.Now())
var byteSliceType reflect.Type = reflect.TypeOf([]byte{})

func (r *SelectResult) Next() (isNext bool, err error) {
	val := reflect.ValueOf(r.query.getRoot().structPtr).Elem()
	cs := make([]interface{}, 0, val.NumField())
	for i := 0; i < val.NumField(); i++ {
		valFld := val.Field(i)
		cs = append(cs, valFld.Addr().Interface())
	}
	isNext = r.rows.Next()
	if !isNext {
		return
	}
	err = r.rows.Err()
	if err != nil {
		return
	}
	err = r.rows.Scan(cs...)
	if err != nil {
		return
	}
	return
}

var capPat *regexp.Regexp = regexp.MustCompile(`[A-Z]+[a-z]*|[a-z]+`)

func capWordsToSnakeCase(name string) string {
	matches := capPat.FindAllString(name, -1)
	s := ""
	for i, m := range matches {
		if i > 0 {
			s += "_"
		}
		s += strings.ToLower(m)
	}
	return s
}

func isSerial(f reflect.StructField) (result bool) {
	result, _ = strconv.ParseBool(f.Tag.Get("db-is-serial"))
	return
}

func columnName(f reflect.StructField) string {
	return capWordsToSnakeCase(f.Name)
}

func fieldAuto(f reflect.StructField) string {
	for _, key := range []string{"db-on-insert", "db-on-update"} {
		s := f.Tag.Get(key)
		if s != "" {
			return s
		}
	}
	return ""
}

func safeParseBool(s string, def bool) bool {
	ret, err := strconv.ParseBool(s)
	if err != nil {
		return def
	} else {
		return ret
	}
}

func compileInsert(relation string, structPtr interface{}) (sql string, bind []interface{}) {
	val := reflect.ValueOf(structPtr).Elem()
	typ := val.Type()

	sql = "INSERT INTO "
	sql += `"` + relation + `" `
	sql += "("
	numFields := 0
	skip := make(map[int]bool)
	for i := 0; i < typ.NumField(); i += 1 {
		fld := typ.Field(i)
		if isSerial(fld) {
			skip[i] = true
			continue
		}
		if safeParseBool(fld.Tag.Get("db-zero-is-null"), false) {
			z := reflect.Zero(val.Field(i).Type())
			if reflect.DeepEqual(z.Interface(), val.Field(i).Interface()) {
				skip[i] = true
				continue
			}
		}
		if numFields > 0 {
			sql += ", "
		}
		sql += fmt.Sprintf(`"%s"`, columnName(fld))
		numFields += 1
	}
	sql += ")"
	sql += "VALUES ("
	numFields = 0
	for i := 0; i < typ.NumField(); i += 1 {
		if skip[i] {
			continue
		}
		fld := typ.Field(i)
		valFld := val.Field(i)

		if numFields > 0 {
			sql += ", "
		}
		numFields += 1
		onInsert := fld.Tag.Get("db-on-insert")
		onUpdate := fld.Tag.Get("db-on-update")
		if onInsert != "" {
			sql += onInsert
		} else if onUpdate != "" {
			sql += onUpdate
		} else {
			bind = append(bind, valFld.Interface())
			sql += fmt.Sprintf("$%d", len(bind))
		}
	}
	sql += ")"
	return
}

func Insert(db *sql.DB, relation string, structPtr interface{}) (err error) {
	if db == nil {
		db, err = DBConn()
		if err != nil {
			return
		}
		defer db.Close()
	}

	if reflect.ValueOf(structPtr).Kind() != reflect.Ptr {
		err = fmt.Errorf("expecting pointer but got %q", reflect.ValueOf(structPtr).Kind())
		return
	}
	sql, bind := compileInsert(relation, structPtr)
	log.Debug("insert sql: %q", sql)
	_, err = db.Exec(sql, bind...)
	if err != nil {
		return
	}
	val := reflect.ValueOf(structPtr).Elem()
	typ := val.Type()
	for i := 0; i < typ.NumField(); i += 1 {
		typFld := typ.Field(i)
		valFld := val.Field(i)
		if isSerial(typFld) {
			log.Debug("querying serial: %q, %q", relation, columnName(typFld))
			row := db.QueryRow(`SELECT CURRVAL(pg_get_serial_sequence($1, $2))`, relation, columnName(typFld))
			err = row.Scan(valFld.Addr().Interface())
			if err != nil {
				return
			}
		}
	}
	return
}

func compileUpdate(relation string, structPtr interface{}) (sql string, bind []interface{}) {
	sql = fmt.Sprintf(`UPDATE "%s" SET `, relation)
	val := reflect.ValueOf(structPtr).Elem()
	typ := val.Type()
	numFields := 0
	for i := 0; i < val.NumField(); i += 1 {
		typFld := typ.Field(i)
		valFld := val.Field(i)
		if isKey(typFld) {
			continue
		}
		if numFields > 0 {
			sql += ", "
		}
		sql += columnName(typFld)
		sql += " = "
		onUpdate := typFld.Tag.Get("db-on-update")
		if onUpdate != "" {
			sql += onUpdate
		} else {
			bind = append(bind, valFld.Interface())
			sql += fmt.Sprintf("$%d", len(bind))
		}
		numFields += 1
	}
	sql += " WHERE "
	numFields = 0
	for i := 0; i < val.NumField(); i += 1 {
		typFld := typ.Field(i)
		valFld := val.Field(i)
		if !isKey(typFld) {
			continue
		}
		if numFields > 0 {
			sql += " AND "
		}
		numFields += 1
		bind = append(bind, valFld.Interface())
		sql += fmt.Sprintf(`"%s" = $%d`, columnName(typFld), len(bind))
	}
	return
}

func Update(db *sql.DB, relation string, structPtr interface{}) (err error) {
	if db == nil {
		db, err = DBConn()
		if err != nil {
			return
		}
		defer db.Close()
	}
	sql, bind := compileUpdate(relation, structPtr)
	log.Debug("update sql: %q", sql)
	_, err = db.Exec(sql, bind...)
	return
}

func isKey(f reflect.StructField) bool {
	if isSerial(f) {
		return true
	}
	t := f.Tag.Get("db-is-key")
	ret, _ := strconv.ParseBool(t)
	return ret
}

func init() {
}
