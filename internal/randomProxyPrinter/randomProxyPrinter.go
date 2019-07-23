package randomProxyPrinter

import (
	"context"
	"database/sql"
	log "github.com/sirupsen/logrus"
)

type RandomProxyPrinter struct {
	logEntry              *log.Entry
	db                    *sql.DB
	incrementValueChannel <-chan interface{}
	decrementValueChannel <-chan interface{}
	printCardChannel      <-chan interface{}
	outputValueChannel    chan<- int
	outputProxyChannel    chan<- Proxy
	value                 int
}

func NewRandomProxyPrinter(logger *log.Logger,
	db *sql.DB,
	incrementValueChannel <-chan interface{},
	decrementValueChannel <-chan interface{},
	printCardChannel <-chan interface{},
	outputValueChannel chan<- int,
	outputProxyChannel chan<- Proxy) *RandomProxyPrinter {
	randomProxyPrinter := &RandomProxyPrinter{
		logEntry:              log.NewEntry(logger),
		db:                    db,
		incrementValueChannel: incrementValueChannel,
		decrementValueChannel: decrementValueChannel,
		printCardChannel:      printCardChannel,
		outputValueChannel:    outputValueChannel,
		outputProxyChannel:    outputProxyChannel,
	}

	return randomProxyPrinter
}

func (t *RandomProxyPrinter) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.incrementValueChannel:
			t.value++

			if t.value == 14 {
				t.value++
			}

			if t.value > 16 {
				t.value = 0
			}

			t.logEntry.
				WithField("value", t.value).
				Trace("incremented value")

			t.outputValueChannel <- t.value
		case <-t.decrementValueChannel:
			t.value--

			if t.value == 14 {
				t.value--
			}

			if t.value < 0 {
				t.value = 16
			}

			t.logEntry.
				WithField("value", t.value).
				Trace("decremented value")

			t.outputValueChannel <- t.value
		case <-t.printCardChannel:
			logEntry := t.logEntry.
				WithField("value", t.value)

			logEntry.Trace("fetching random proxy from database")

			proxy := Proxy{}

			row := t.db.QueryRow("SELECT name, description, illustration FROM proxies WHERE value = ? ORDER BY RANDOM() LIMIT 1", t.value)

			err := row.Scan(&proxy.Name,
				&proxy.Description,
				&proxy.Illustration)

			if err != nil {
				logEntry.WithError(err).
					Error("failed to fetch random proxy from database")

				return err
			}

			logEntry.WithField("proxy", proxy).
				Trace("random proxy fetched from database")

			t.outputProxyChannel <- proxy
		}
	}
}
