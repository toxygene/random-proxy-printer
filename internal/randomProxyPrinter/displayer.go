package randomProxyPrinter

type Displayer interface {
	Display(Proxy) error
}
