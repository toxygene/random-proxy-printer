package randomProxyPrinter

import (
	"context"
	"database/sql"
)

type RandomProxyPrinter struct {
	db					  *sql.DB
	incrementValueChannel <-chan interface {}
	decrementValueChannel <-chan interface {}
	printButtonChannel    <-chan interface {}
	displayChannel        chan<- int
	printChannel          chan<- Proxy
	value                 int
}

func NewRandomProxyPrinter(db *sql.DB,
	incrementValueChannel <-chan interface {},
	decrementValueChannel <-chan interface {},
	printButtonChannel <-chan interface {},
	displayChannel chan<- int,
	printChannel chan<- Proxy) *RandomProxyPrinter {
	randomProxyPrinter := &RandomProxyPrinter{
		db:                    db,
		incrementValueChannel: incrementValueChannel,
		decrementValueChannel: decrementValueChannel,
		printButtonChannel:    printButtonChannel,
		displayChannel:        displayChannel,
		printChannel:          printChannel,
	}

	return randomProxyPrinter
}

func(t *RandomProxyPrinter) Run(ctx context.Context) error {
	for {
		select {
		case <- ctx.Done():
			return ctx.Err()
		case <- t.incrementValueChannel:
			t.value++

			if t.value == 14 {
				t.value++
			}

			if t.value > 16 {
				t.value = 0
			}

			t.displayChannel <- t.value
		case <- t.decrementValueChannel:
			t.value--

			if t.value == 14 {
				t.value--
			}

			if t.value < 0 {
				t.value = 16
			}

			t.displayChannel <- t.value

		case <- t.printButtonChannel:
			proxy := Proxy{}

			row := t.db.QueryRow("SELECT name, description, illustration FROM proxies WHERE value = ? ORDER BY RANDOM() LIMIT 1", t.value)

			err := row.Scan(&proxy.Name,
				&proxy.Description,
				&proxy.Illustration)

			if err != nil {
				return err
			}

			t.printChannel <- proxy
		}
	}
}