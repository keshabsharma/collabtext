package server

import (
	t "collabtext/transform"
	"sync"
)

type Document struct {
	content    string
	revision   uint64
	transforms []t.Operation

	opLock sync.Mutex
}

func (d *Document) ProcessTransformation(ot *t.Operation) (*t.Operation, error) {
	d.opLock.Lock()
	defer d.opLock.Unlock()

	// apply operation to the document
	// add to transforms
	// update revision
	ot.Revision++

	return ot, nil
}

func (d *Document) ApplyTransformations(ot []*t.Operation) error {
	d.opLock.Lock()
	defer d.opLock.Unlock()

	// get list of transformations and just apply

	return nil
}
