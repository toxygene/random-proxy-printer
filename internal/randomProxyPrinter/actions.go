package randomProxyPrinter

type Action string

const (
	IncrementValue   Action = "increment"
	DecrementValue          = "decrement"
	PrintRandomProxy        = "print"
)
