package randomProxyPrinter

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type RandomProxyPrinter struct {
	db        *sqlx.DB
	displayer Displayer
	inputter  Inputter
	logger    *logrus.Entry
	printer   Printer
}

func NewRandomProxyPrinter(db *sqlx.DB,
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

func (t *RandomProxyPrinter) GetDistinctValues() ([]int, error) {
	var distinctValues []int

	if err := t.db.Select(&distinctValues, "select distinct value from proxies order by length(value), value"); err != nil {
		t.logger.Error("get distinct values: %w", err)
		return nil, fmt.Errorf("get distinct values: %w", err)
	}

	return distinctValues, nil
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

		offset := 0

		distinctValues, err := t.GetDistinctValues()
		if err != nil {
			t.logger.Errorf("run proxy printer: %v", err)
			return fmt.Errorf("run proxy printer: %w", err)
		}

		for action := range actions {
			t.logger.WithField("action", action).Trace("read action")

			if action == IncrementValue {
				offset = euclideanModulo(offset+1, len(distinctValues))

				t.logger.WithField("offset", offset).WithField("value", distinctValues[offset]).Trace("incremented offset")

				if err := t.displayer.Display(distinctValues[offset]); err != nil {
					return fmt.Errorf("increment display value: %w", err)
				}
			} else if action == DecrementValue {
				offset = euclideanModulo(offset-1, len(distinctValues))

				t.logger.WithField("offset", offset).WithField("value", distinctValues[offset]).Trace("decremented offset")

				if err := t.displayer.Display(distinctValues[offset]); err != nil {
					return fmt.Errorf("decrement display value: %w", err)
				}
			} else if action == PrintRandomProxy {
				logEntry := t.logger.
					WithField("value", distinctValues[offset])

				logEntry.Trace("fetching random proxy from database")

				proxy := Proxy{}

				row := t.db.QueryRow("SELECT name, print_data FROM proxies WHERE value = ? ORDER BY RANDOM() LIMIT 1", distinctValues[offset])

				if err := row.Scan(&proxy.Name, &proxy.PrintData); err != nil {
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

func euclideanModulo(a int, b int) int {
	return (a%b + b) % b
}
