package server

import (
	t "collabtext/transform"
	"fmt"
	"log"
	"sync"
)

type Document struct {
	name       string
	content    string
	revision   uint64
	transforms []t.Operation

	opLock sync.Mutex
}

func newDocument(doc string) *Document {
	d := &Document{
		content:    "test",
		revision:   0,
		transforms: make([]t.Operation, 0),
		name:       doc,
	}
	op := &t.Operation{0, "insert", 0, "", "server", doc, ""}
	d.revision = 0
	d.transforms = append(d.transforms, *op)
	return d
}

func (d *Document) applyOp(op *t.Operation) {
	document := d.content
	if op.Op == "insert" {
		if op.Position >= int32(len(document)) {
			document = document + op.Str
		} else {
			document = document[0:op.Position] + op.Str + document[op.Position:]
		}
	} else {
		if op.Position >= int32(len(document)) {
			document = document[0 : op.Position-1]
		} else {
			document = document[0:op.Position] + document[op.Position+1:]
		}
	}
	log.Println(document)
	d.content = document
}

func (d *Document) ProcessTransformation(ot *t.Operation) (*t.Operation, error) {
	d.opLock.Lock()
	defer d.opLock.Unlock()

	if d.revision+1 != ot.Revision {
		return nil, fmt.Errorf("invalid revision no")
	}

	// apply operation to the document
	// add to transforms
	// update revision
	d.revision++
	d.transforms = append(d.transforms, *ot)

	return ot, nil
}

func (d *Document) ApplyTransformations(ops []t.Operation) error {
	d.opLock.Lock()
	defer d.opLock.Unlock()

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
	return nil
}

func (d *Document) GetTransformations(from uint64) []t.Operation {
	// for i, v := range d.transforms {
	// 	if v.Revision == from {
	// 		return d.transforms[i:]
	// 	}
	// }
	// return make([]t.Operation, 0)
	return d.transforms[from:]
}
