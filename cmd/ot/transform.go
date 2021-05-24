package main

import (
	"fmt"
)


type Operation struct {
	// revision
	rev int32

	// insert or delete
	op string

	// position
	pos int32

	// string
	str string

	// origin
	client string
}

/*
Tii(Ins[p1,c1], Ins[p2, c2]) {
      if p1 < p2  or (p1 = p2 and u1 > u2) // breaking insert-tie using user identifiers (u1, u2)
            return Ins[p1, c1];  // e.g. Tii(Ins[3, “a”], Ins[4, “b”]) = Ins[3, “a”]
      else return Ins[p1+1, c1]; } // Tii(Ins[3, “a”], Ins[1, “b”]) = Ins[4, “a”]
 
Tid(Ins[p1,c1], Del[p2]) {          
      if p1 <= p2 return Ins[p1, c1]; // e.g. Tid(Ins[3, “a”], Del[4]) = Ins[3, “a”]
     else return Ins[p1-1, c1]; } // Tid(Ins[3, “a”], Del[1] ) = Ins[2, “a”]
 
Tdi(Del[p1], Ins[p2, c2]) {
      if p1 < p2 return Del[p1];  // e.g.  Tdi(Del[3], Ins[4, “b”]) = Del[3]
      else return Del[p1+1]; } // Tdi(Del[3], Ins[1, “b”]) = Del[4]
 
Tdd(Del[p1], Del[p2]) {
      if p1 < p2 return Del[p1]; // e.g.   Tdd(Del[3], Del[4]) = Del[3]
      else if p1 > p2 return Del[p1-1]; // Tdd(Del[3], Del[1]) = Del[2]
      else return I; } // breaking delete-tie using I (identity op)  Tdd(Del[3]. Del[3]) = I

*/

func transformInsertInsert(op1, op2 Operation) Operation {
	if op1.pos < op2.pos || (op1.pos == op2.pos && op1.rev > op2.rev) {
		return op1
	} else {
		op1.pos++
		return op1;
	}
}


func transformInsertDelete(op1, op2 Operation) Operation {
	if op1.pos <= op2.pos {
		return op1
	} else {
		op1.pos--
		return op1;
	}
}

func transformDeleteInsert(op1, op2 Operation) Operation {
	if op1.pos < op2.pos {
		return op1
	} else {
		op1.pos++
		return op1;
	}
}

func transformDeleteDelete(op1, op2 Operation) Operation {
	if op1.pos < op2.pos {
		return op1
	} else if op1.pos > op2.pos {
		return op2
	} else {
		//return nil
		op1.pos = 0
		op1.str = ""
		return op1
	}
}

func transform(op1, op2 Operation) Operation {
	if op1.op == "insert" && op2.op == "insert" {
		return transformInsertInsert(op1,op2)
	}
	if op1.op == "insert" && op2.op == "delete" {
		return transformInsertDelete(op1,op2)
	}
	if op1.op == "delete" && op2.op == "insert" {
		return transformDeleteInsert(op1,op2)
	}
	if op1.op == "delete" && op2.op == "delete" {
		return transformDeleteDelete(op1, op2)
	}
	return Operation{}
}

func getOp(client string, rev int32, op string, str string, pos int32) Operation {
	return Operation{rev, op, pos, str, client}
}

func applyOp(document string, op Operation) string {
	if op.op == "insert" {
		return document[0:op.pos] + op.str + document[op.pos:]
	} else {
		return document[0:op.pos] + document[op.pos+1:]
	}
}

//var serverRev = 0
func main() {
	revlog := make([]Operation, 10)
	var revision int32 = 1
	document := ""
	lastOp := getOp("A", revision, "insert", "A", 0)
	document = applyOp(document, lastOp)
	//revlog[revision]
	revision++
	fmt.Println(document)

	nextOp := getOp("C", revision, "insert", "C", 0)
	nextOp = transform(nextOp, revlog[revision - 1])
	document = applyOp(document, nextOp)
	revlog[revision] = nextOp
	revision++
	fmt.Println(document)

	Server:
			 b 1 0
	a 1 2 -> b 1 3 - stop
	c 2 0
	d 3 1
	b 1 3
	
	

	cdabhello
	-- aplied
	x 1 0
	
	-- unapplied
	a 2 0  -> b 1 1 -> stop
	c 3 0  ->  b 1 2
	d 3 10 -> b 1 2

	b 2 0 -> b 1 3

	cdab



	bcd(b)ahello
	a 1 0 -> b 1 1
	c 2 0 -> b 1 2
	d 3 1 -> b 
	
	b 1 0 -> b r 3

	a 1 0 --> b 1 2
	c 2 0 --> b 1 1

	a 1 0 --> b 1 3
	c 2 0 --> b 1 2
	d 3 1 --> b 1 1

	a
	ca
	cda

	b
	ab
	cab
	cdab
	
	lateOp := getOp("B", 1, "insert", "B", 0)
	if lateOp.rev < revision {
		lateOp = transform(lateOp, revlog[revision - 1])
	}
	document = applyOp(document, lateOp)
	revlog[revision] = lateOp
	revision++

	fmt.Println(document)

}
