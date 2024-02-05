package httpjson

import (
	"fmt"
)

type ConnectionInfo struct {
	Host     string
	Port     int
	Password string
}

func (c ConnectionInfo) String() string {
	return fmt.Sprintf("http://%s:%d\nPassword: %s", c.Host, c.Port, c.Password)
}
