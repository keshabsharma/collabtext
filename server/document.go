package server

import (
	t "collabtext/transform"
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

func (d *Document) ProcessTransformation(ot *t.Operation) (*t.Operation, error) {
	d.opLock.Lock()
	defer d.opLock.Unlock()

	op_t := performOp(d.transforms, *ot)
	d.applyOp(op_t)
	d.revision++
	ot.Revision = d.revision
	d.transforms = append(d.transforms, op_t)

	log.Println("Server Doc: ", d.content)

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
			d.applyOp(op)
			d.transforms = append(d.transforms, op)
			d.revision = op.Revision
		}
	}
	return nil
}

func (d *Document) GetTransformations(from uint64) []t.Operation {
	return d.transforms[from:]
}

// TRANSFORM

func transformInsertInsert(op1, op2 t.Operation) t.Operation {
	if op1.Position < op2.Position || (op1.Position == op2.Position && op1.Revision > op2.Revision) {
		return op1
	} else {
		op1.Position++
		return op1
	}
}

func transformInsertDelete(op1, op2 t.Operation) t.Operation {
	if op1.Position <= op2.Position {
		return op1
	} else {
		op1.Position--
		return op1
	}
}

func transformDeleteInsert(op1, op2 t.Operation) t.Operation {
	if op1.Position < op2.Position {
		return op1
	} else {
		op1.Position++
		return op1
	}
}

func transformDeleteDelete(op1, op2 t.Operation) t.Operation {
	if op1.Position < op2.Position {
		return op1
	} else if op1.Position > op2.Position {
		op1.Position--
		return op1
	} else {
		//fmt.Println("huh")
		op1.Position = 0
		op1.Str = ""
		op1.Op = "insert"
		return op1
	}
}

func transform(op1, op2 t.Operation) t.Operation {
	if op1.Op == "insert" && op2.Op == "insert" {
		return transformInsertInsert(op1, op2)
	}
	if op1.Op == "insert" && op2.Op == "delete" {
		return transformInsertDelete(op1, op2)
	}
	if op1.Op == "delete" && op2.Op == "insert" {
		return transformDeleteInsert(op1, op2)
	}
	if op1.Op == "delete" && op2.Op == "delete" {
		return transformDeleteDelete(op1, op2)
	}
	return t.Operation{}
}

func (d *Document) applyOp(op t.Operation) {
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
	d.content = document
}

func getRefOp(opList []t.Operation, op t.Operation) []t.Operation {
	return opList[op.Revision:]
}

func performOp(revlog []t.Operation, op t.Operation) t.Operation {
	refOps := getRefOp(revlog, op)
	for _, rOp := range refOps {
		op = transform(op, rOp)
	}
	return op
}
