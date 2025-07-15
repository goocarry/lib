package clickhouselogger

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const (
	DefaultBatch          = 10000
	DefaultConsumeTimeout = 5      // seconds
	DefaultBuffer         = 100000 // how many entries to keep
)

type Log struct {
	Date time.Time
	Time time.Time
	Body interface{}
}

type Opt func(instance *ClickHouseLogger) *ClickHouseLogger

func WithBatchSize(size int) Opt {
	return func(instance *ClickHouseLogger) *ClickHouseLogger {
		instance.batchSize = size
		return instance
	}
}

func WithConsumePeriod(seconds uint) Opt {
	return func(instance *ClickHouseLogger) *ClickHouseLogger {
		instance.consumeTimeoutSeconds = seconds
		return instance
	}
}

type ClickHouseLogger struct {
	loggerName    string
	eventsChannel chan Log
	sigChannel    chan bool

	clickhouseClient      *sql.DB
	tableName             string
	tableColumns          []string
	batchSize             int
	consumeTimeoutSeconds uint
}

// TODO
// handle errors
// use general body payload
func New(loggerName string, tableName string, tableColumns []string, chClient *sql.DB, opts ...Opt) *ClickHouseLogger {
	instance := &ClickHouseLogger{
		loggerName:            loggerName,
		tableName:             tableName,
		tableColumns:          tableColumns,
		consumeTimeoutSeconds: DefaultConsumeTimeout,
		clickhouseClient:      chClient,
		eventsChannel:         make(chan Log, DefaultBuffer),
		batchSize:             DefaultBatch,
	}

	for i := range opts {
		instance = opts[i](instance)
	}

	return instance
}

func (s *ClickHouseLogger) BalanceChangeMessage(l Log) {
	select {
	case s.eventsChannel <- l:
		return
	default:

	}
}

func (s *ClickHouseLogger) Stop() {
	s.sigChannel <- true
}

func (s *ClickHouseLogger) Consumer() {
	for {
		select {
		case _ = <-s.sigChannel:
			close(s.eventsChannel)
			return
		default:
		}

		func() {
			time.Sleep(time.Second * time.Duration(s.consumeTimeoutSeconds))

			chanLen := len(s.eventsChannel)

			if chanLen < 1 {
				return
			}
			messages := make([]Log, 0)
			for i := 0; i < chanLen; i++ {
				select {
				case msg := <-s.eventsChannel:
					messages = append(messages, msg)
				default:
					continue
				}
			}

			tx, err := s.clickhouseClient.Begin()
			if err != nil {
				return
			}
			stmt, err := tx.Prepare(buildInsertQuery(s.tableName, s.tableColumns))

			for i := 0; i < len(messages); i += s.batchSize {
				batch := messages[i:checkMin(i+s.batchSize, len(messages))]
				for j := 0; j < len(batch); j++ {

					rec := batch[j]
					_, err := stmt.Exec(
						rec.Date,
						rec.Time,
					)
					if err != nil {
						return
					}
				}
			}

			if err := tx.Commit(); err != nil {
			}
		}()
	}
}

func checkMin(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func buildInsertQuery(tableName string, columns []string) string {
	var ph []string
	for i := 0; i < len(columns); i++ {
		ph = append(ph, "?")
	}
	return fmt.Sprintf("insert into %s (%s) values (%s)", tableName, strings.Join(columns, ","), strings.Join(ph, ","))
}
