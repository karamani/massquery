package main

import "database/sql"

type scanContainer struct {
	Pointers []interface{}
	Values   []sql.RawBytes
}

func newScanContainer(size int) *scanContainer {
	c := &scanContainer{}
	if size > 0 {
		c.Pointers = make([]interface{}, size)
		c.Values = make([]sql.RawBytes, size)
		for i := range c.Pointers {
			c.Pointers[i] = &c.Values[i]
		}
	}
	return c
}

func (c *scanContainer) AsStrings() []string {
	strings := make([]string, len(c.Values))
	for i, elem := range c.Values {
		strings[i] = ""
		if elem != nil {
			strings[i] = string(elem)
		}
	}
	return strings
}
