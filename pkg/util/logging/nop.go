package logging

type nopLogger struct {
}

func NewNopLogger() Logger {
	return nopLogger{}
}

func (l nopLogger) Info(_ string, _ ...any) {

}

func (l nopLogger) Err(_ string, _ ...any) {

}

func (l nopLogger) WithPrefix(_ string) Logger {
	return nopLogger{}
}
