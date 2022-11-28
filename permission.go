package main

import (
	"fmt"
)

type Permission struct {
	ID        int
	Namespace string
	Action    string
}

func (p *Permission) String() string {
	return fmt.Sprintf("%s:*#%s", p.Namespace, p.Action)
}
