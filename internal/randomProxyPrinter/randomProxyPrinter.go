package randomProxyPrinter

import (
	"context"
	"database/sql"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var StopRunningError = errors.New("stop running error")

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
		case _, ok := <-t.incrementValueChannel:
			if !ok {
				return errors.New("increment value channel unexpectedly closed")
			}

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
		case _, ok := <-t.decrementValueChannel:
			if !ok {
				return errors.New("decrement value channel unexpectedly closed")
			}

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
		case _, ok := <-t.printCardChannel:
			if !ok {
				return errors.New("print card channel unexpectedly closed")
			}

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
