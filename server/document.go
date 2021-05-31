package server

import (
	t "collabtext/transform"
	"log"
	"sync"
)

type Document struct {
	content    string
	revision   uint64
	transforms []t.Operation

	opLock sync.Mutex
}

func newDocument() *Document {
	return &Document{
		content:    "",
		revision:   1,
		transforms: make([]t.Operation, 0),
	}
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

func (d *Document) ApplyTransformations(ops []t.Operation) error {
	d.opLock.Lock()
	defer d.opLock.Unlock()

	log.Println("applying transforms ")
	for _, v := range ops {
		log.Println(v)
	}

	for _, op := range ops {
		if op.Revision <= d.revision {
			continue
		}

		// making sure all the transforms are in order
		if d.revision+1 == op.Revision {
			// apply
			d.transforms = append(d.transforms, op)
			d.revision = op.Revision
		}
	}

	// get list of transformations and just apply

	return nil
}

func (d *Document) GetTransformations(from uint64) []t.Operation {

	return d.transforms[from:]
}
