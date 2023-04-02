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
	t.logger.Info("proxy printer starting")
	defer t.logger.Info("proxy printer ending")

	if err := t.displayer.Display(0); err != nil {
		return fmt.Errorf("initialize display to 0: %w", err)
	}

	g, ctx := errgroup.WithContext(parentCtx)

	actions := make(chan Action)

	g.Go(func() error {
		defer close(actions)

		t.logger.Info("inputter goroutine started")
		defer t.logger.Info("inputter goroutine finished")

		if err := t.inputter.Run(ctx, actions); err != nil {
			return fmt.Errorf("run inputter: %w", err)
		}

		return nil
	})

	g.Go(func() error {
		t.logger.Info("actions handler goroutine started")
		defer t.logger.Info("actions handler goroutine finished")

		for action := range actions {
			t.logger.WithField("action", action).Trace("read action")

			if action == IncrementValue {
				t.value++

				if t.value == 14 {
					t.value++
				}

				if t.value > 16 {
					t.value = 0
				}

				t.logger.WithField("value", t.value).Trace("incremented value")

				if err := t.displayer.Display(t.value); err != nil {
					return fmt.Errorf("increment display value: %w", err)
				}
			} else if action == DecrementValue {
				t.value--

				if t.value == 14 {
					t.value--
				}

				if t.value < 0 {
					t.value = 16
				}

				t.logger.WithField("value", t.value).Trace("decremented value")

				if err := t.displayer.Display(t.value); err != nil {
					return fmt.Errorf("decrement display value: %w", err)
				}
			} else if action == PrintRandomProxy {
				logEntry := t.logger.
					WithField("value", t.value)

				logEntry.Trace("fetching random proxy from database")

				proxy := Proxy{}

				row := t.db.QueryRow("SELECT name, description, print_data FROM proxies WHERE value = ? ORDER BY RANDOM() LIMIT 1", t.value)

				if err := row.Scan(&proxy.Name, &proxy.Description, &proxy.PrintData); err != nil {
					logEntry.WithError(err).Error("failed to fetch random proxy from database")

					return err
				}

				logEntry.WithField("proxy_name", proxy.Name).Trace("random proxy fetched from database")

				if err := t.printer.Print(proxy); err != nil {
					logEntry.WithError(err).Error("failed to print random proxy")

					return err
				}
			}
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("run proxy printer groups: %w", err)
	}

	return nil
}
