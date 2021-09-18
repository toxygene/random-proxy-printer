package randomProxyPrinter

type Action string

const (
    IncrementValue   Action = "increment"
    DecrementValue   Action = "decrement"
    PrintRandomProxy Action = "print"
)
