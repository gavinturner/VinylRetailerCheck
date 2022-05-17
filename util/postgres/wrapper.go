package postgres

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"

	"github.com/gavinturner/vinylretailers/util/log"
)

const (
	ENV_VAR_NAME_DISABLED             = "LONG_QUERY_LOGGING_OFF"
	ENV_VAR_NAME_THRESHOLD            = "LONG_QUERY_THRESHOLD_SECS"
	DEFAULT_LONG_QUERY_THRESHOLD_SECS = 10
)

var (
	LongQueryThresholdSecs int
	LoggingDisabled        bool
	Logger                 *logrus.Entry
)

func init() {
	v := os.Getenv(ENV_VAR_NAME_THRESHOLD)
	if strings.TrimSpace(v) != "" {
		var err error
		LongQueryThresholdSecs, err = strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			log.Errorf("LONG_QUERY_THTRESHOLD_SECS is set to unusable value '%v'", strings.TrimSpace(v))
			LongQueryThresholdSecs = DEFAULT_LONG_QUERY_THRESHOLD_SECS
		}
	} else {
		LongQueryThresholdSecs = DEFAULT_LONG_QUERY_THRESHOLD_SECS
	}
	v = os.Getenv(ENV_VAR_NAME_DISABLED)
	if strings.TrimSpace(v) != "" {
		LoggingDisabled = strings.ToLower(strings.TrimSpace(v)) == "true" || strings.ToLower(strings.TrimSpace(v)) == "yes"
	} else {
		LoggingDisabled = false
	}
}

// Tx exposes all the methods available on the *sqlx.Tx struct
type Tx struct {
	*sqlx.Tx
}

func (db DB) Beginx() (*Tx, error) {
	tx, err := db.DB.Beginx()
	if err != nil {
		return nil, err
	}
	return &Tx{Tx: tx}, err
}

func (db DB) Begin() (*Tx, error) {
	return db.Beginx()
}

func (db DB) Select(dest interface{}, query string, args ...interface{}) error {
	timer := NewTimer("Select", query, args)
	e := db.DB.Select(dest, query, args...)
	timer.Stop()
	return e
}

func (db DB) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	timer := NewTimer("Queryx", query, args)
	r, e := db.DB.Queryx(query, args...)
	timer.Stop()
	return r, e
}

func (db DB) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	timer := NewTimer("QueryRowx", query, args)
	r := db.DB.QueryRowx(query, args...)
	timer.Stop()
	return r
}

func (db DB) NamedExec(query string, arg interface{}) (sql.Result, error) {
	timer := NewTimer("NamedExec", query, []interface{}{arg})
	r, e := db.DB.NamedExec(query, arg)
	timer.Stop()
	return r, e
}

func (db DB) Get(dest interface{}, query string, args ...interface{}) error {
	timer := NewTimer("Get", query, args)
	e := db.DB.Get(dest, query, args...)
	timer.Stop()
	return e
}

func (db DB) QueryRow(query string, args ...interface{}) *sql.Row {
	timer := NewTimer("QueryRow", query, args)
	r := db.DB.QueryRow(query, args...)
	timer.Stop()
	return r
}

func (db DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	timer := NewTimer("Query", query, args)
	r, e := db.DB.Query(query, args...)
	timer.Stop()
	return r, e
}

func (tx Tx) NamedExec(query string, arg interface{}) (sql.Result, error) {
	timer := NewTimer("Tx.NamedExec", query, []interface{}{arg})
	r, e := tx.Tx.NamedExec(query, arg)
	timer.Stop()
	return r, e
}

func (tx Tx) Select(dest interface{}, query string, args ...interface{}) error {
	timer := NewTimer("Tx.Select", query, args)
	e := tx.Tx.Select(dest, query, args...)
	timer.Stop()
	return e
}

func (tx Tx) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	timer := NewTimer("Tx.Queryx", query, args)
	r, e := tx.Tx.Queryx(query, args...)
	timer.Stop()
	return r, e
}

func (tx Tx) Get(dest interface{}, query string, args ...interface{}) error {
	timer := NewTimer("Tx.Get", query, args)
	e := tx.Tx.Get(dest, query, args...)
	timer.Stop()
	return e
}

func (tx Tx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	timer := NewTimer("Tx.Query", query, args)
	r, e := tx.Tx.Query(query, args...)
	timer.Stop()
	return r, e
}

func (tx Tx) QueryRow(query string, args ...interface{}) *sql.Row {
	timer := NewTimer("Tx.QueryRow", query, args)
	r := tx.Tx.QueryRow(query, args...)
	timer.Stop()
	return r
}

func (tx Tx) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	timer := NewTimer("Tx.QueryRowx", query, args)
	r := tx.Tx.QueryRowx(query, args...)
	timer.Stop()
	return r
}

//
// Encapsulate a timer log request for the	 DB query wrapper
//

type Timer struct {
	Start       time.Time
	FuncName    string
	QueryString string
	Args        []interface{}
}

func NewTimer(funcName string, query string, args []interface{}) *Timer {
	return &Timer{
		Start:       time.Now(),
		FuncName:    funcName,
		QueryString: query,
		Args:        args,
	}
}

func (t *Timer) Stop() {
	if !LoggingDisabled {
		secs := time.Since(t.Start).Seconds()
		if int(secs) > LongQueryThresholdSecs {
			if len(t.QueryString) > 500 {
				t.QueryString = t.QueryString[0:450] + " ... " + t.QueryString[len(t.QueryString)-50:]
			}
			argsStr := formatArgsToString(t.Args)
			if len(argsStr) > 500 {
				argsStr = argsStr[0:450] + " ... " + argsStr[len(t.QueryString)-50:]
			}
			log.New().WithFields(logrus.Fields{
				"seconds":  secs,
				"funcName": t.FuncName,
				"query":    t.QueryString,
				"args":     argsStr,
			}).Warn("LONG QUERY")
		}
	}
}

// Just output a comma separated list of the string representation of each arg
func formatArgsToString(args []interface{}) string {
	str := &strings.Builder{}
	for i := 0; i < len(args); i++ {
		if i != 0 {
			str.WriteString(", ")
		}

		if argAsStr, ok := args[i].(string); ok {
			str.WriteString("'")
			str.WriteString(argAsStr)
			str.WriteString("'")
		} else {
			_, _ = fmt.Fprintf(str, "%v", args[i])
		}
	}

	return str.String()
}
