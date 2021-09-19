package randomProxyPrinter

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type RandomProxyPrinter struct {
	db        *sql.DB
	displayer Displayer
	inputter  Inputter
	logger    *logrus.Entry
	printer   Printer
	value     int
}

func NewRandomProxyPrinter(db *sql.DB,
	displayer Displayer,
	inputter Inputter,
	printer Printer,
	logger *logrus.Entry) *RandomProxyPrinter {
	randomProxyPrinter := &RandomProxyPrinter{
		db:        db,
		displayer: displayer,
		inputter:  inputter,
		logger:    logger,
		printer:   printer,
	}

	return randomProxyPrinter
}

func (t *RandomProxyPrinter) Run(parentCtx context.Context) error {
	if err := t.displayer.Display(0); err != nil {
		return fmt.Errorf("initialize display to 0: %w", err)
	}

	ctx, cancel := context.WithCancel(parentCtx)

	g := new(errgroup.Group)

	actions := make(chan Action)

	g.Go(func() error {
		defer close(actions)

		return t.inputter.Run(ctx, actions)
	})

	g.Go(func() error {
		for action := range actions {
			if action == IncrementValue {
				t.value++

				if t.value == 14 {
					t.value++
				}

				if t.value > 16 {
					t.value = 0
				}

				t.logger.
					WithField("value", t.value).
					Trace("incremented value")

				t.displayer.Display(t.value)
			} else if action == DecrementValue {
				t.value--

				if t.value == 14 {
					t.value--
				}

				if t.value < 0 {
					t.value = 16
				}

				t.logger.
					WithField("value", t.value).
					Trace("decremented value")

				if err := t.displayer.Display(t.value); err != nil {
					return fmt.Errorf("setting display: %w", err)
				}
			} else if action == PrintRandomProxy {
				logEntry := t.logger.
					WithField("value", t.value)

				logEntry.Trace("fetching random proxy from database")

				proxy := Proxy{}

				row := t.db.QueryRow("SELECT name, description, illustration FROM proxies WHERE value = ? ORDER BY RANDOM() LIMIT 1", t.value)

				if err := row.Scan(&proxy.Name, &proxy.Description, &proxy.Illustration); err != nil {
					logEntry.WithError(err).
						Error("failed to fetch random proxy from database")

					cancel()

					return err
				}

				logEntry.WithField("proxy_name", proxy.Name).
					Trace("random proxy fetched from database")

				if err := t.printer.Print(proxy); err != nil {
					logEntry.WithError(err).
						Error("failed to print random proxy")

					cancel()

					return err
				}
			}
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		t.logger.
			WithError(err).
			Error("run proxy printer failed")

		return fmt.Errorf("run proxy printer groups: %w", err)
	}

	return nil
}
