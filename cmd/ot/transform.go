package main

import (
	. "collabtext/transform"
	"fmt"
)

// type Operation struct {
// 	// revision
// 	rev int32

// 	// insert or delete
// 	op string

// 	// position
// 	pos int32

// 	// string
// 	str string

// 	// origin
// 	client string
// }

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
	if op1.Position < op2.Position || (op1.Position == op2.Position && op1.Revision > op2.Revision) {
		return op1
	} else {
		op1.Position++
		return op1
	}
}

func transformInsertDelete(op1, op2 Operation) Operation {
	if op1.Position <= op2.Position {
		return op1
	} else {
		op1.Position--
		return op1
	}
}

func transformDeleteInsert(op1, op2 Operation) Operation {
	if op1.Position < op2.Position {
		return op1
	} else {
		op1.Position++
		return op1
	}
}

func transformDeleteDelete(op1, op2 Operation) Operation {
	if op1.Position < op2.Position {
		return op1
	} else if op1.Position > op2.Position {
		return op2
	} else {
		op1.Position = 0
		op1.Str = ""
		return op1
	}
}

func transform(op1, op2 Operation) Operation {
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
	return Operation{}
}

func getOp(client string, rev uint64, op string, str string, pos int32) Operation {
	return Operation{rev, op, pos, str, client, "", ""}
}

func applyOp(document string, op Operation) string {
	if op.Op == "insert" {
		return document[0:op.Position] + op.Str + document[op.Position:]
	} else {
		return document[0:op.Position] + document[op.Position+1:]
	}
}

func getRefOp(opList []Operation, op Operation) []Operation {
	return opList[op.Revision:]
}

func getRefOp_L(opList []Operation, op Operation) []Operation {
	refOps := make([]Operation, len(opList))
	var idx int32 = 0
	for _, tmpOp := range opList {
		if tmpOp.Revision >= op.Revision {
			refOps[idx] = tmpOp
			idx++
		}
	}
	return refOps
}

/*
OPlist:
revision 1 : [op1, op2]
revision 2: [op3, op4]
revision 3: [op5, op6,op7]
revison 14



ots:
[o1, o2, o3, o4, o5]

oi (2)
apply

[o1, o2, o3, o4, o5, oi (2,6)]

oj (3)


"eafcdbghello"
		A             			B                   server
1. i a 0; ahello			  ahello
2. i c 1; achello	      	  achello
3. i d 3; achdello
4. i e 0; eachdello
5. i f 2; eafchdello
						  2. i g 2; acghello      i g 4; eafcghdello
						  2. i k 1; akcghello     i k


A						B 								Server
1.i a 0; a												1.i a 0; a
2.i b 1; ab 			1.i c 1; ac 					2.i b 1; ab
					   		   							3.i c 2; abc
2.i d 2; abd											4.i d 3; abcd



For each t in transforms from beginning:
	if t.requestRevision >= ot.requestRevision:
		op = transform(op, t)
or

map[int]requestOts

applied_ot:[..., op1(2,), op2(1,), op3(3,)]


A: i a 0; i c 1; i d 3; i e 0; i f 2;
B:	i g 2 (oi); i k 5 rev6;
C: acbhello -rev3
*/

func getRefOp_LoL(opList [][]Operation, op Operation) []Operation {
	var refOps []Operation
	for _, tmpOps := range opList {
		for _, tOp := range tmpOps {
			if tOp.Revision >= op.Revision {
				refOps = append(refOps, tOp)
			}
		}
	}
	return refOps
}

func performOp(revlog []Operation, op Operation, document string) string {
	refOps := getRefOp_L(revlog, op)
	for _, rOp := range refOps {
		op = transform(op, rOp)
	}
	document = applyOp(document, op)
	return document
}

//var serverRev = 0
// func main() {
// 	revlog := make([]Operation, 10)
// 	var revision uint64 = 1
// 	document := ""
// 	lastOp := getOp("A", revision, "insert", "A", 0)
// 	document = applyOp(document, lastOp)
// 	//revlog[revision]
// 	revision++
// 	fmt.Println(document)

// 	nextOp := getOp("C", revision, "insert", "C", 0)
// 	nextOp = transform(nextOp, revlog[revision-1])
// 	document = applyOp(document, nextOp)
// 	revlog[revision] = nextOp
// 	revision++
// 	fmt.Println(document)

// 	lateOp := getOp("B", 1, "insert", "B", 0)
// 	if lateOp.Revision < revision {
// 		lateOp = transform(lateOp, revlog[revision-1])
// 	}
// 	document = applyOp(document, lateOp)
// 	revlog[revision] = lateOp
// 	revision++

// 	fmt.Println(document)

// }

func normalTest_L() {
	/*
		Clients: clientA, clientB, clientC
		Document: doc, initial with "hello"
		Revision: start with 1
	*/
	revlog := make([]Operation, 10)
	var idx int32 = 0
	document := "hello"

	Op1 := getOp("ClientA", 1, "insert", "Aa", 0) // server: hello ; clientA: Aahello ; clientB: hello ; client C: hello
	document = performOp(revlog, Op1, document)
	revlog[idx] = Op1
	idx++
	fmt.Println(document) // server: Aahello ; clientA: Aahello ; clientB: hello ; client C: hello

	Op2 := getOp("ClientC", 1, "delete", "e", 2) // server: Aahello ; clientA: Aahello ; clientB: hello ; client C: hllo
	document = performOp(revlog, Op1, document)
	revlog[idx] = Op2
	idx++
	fmt.Println(document) // server: Aahllo ; clientA: Aahello ; clientB: hello ; client C: Aahllo

	Op3 := getOp("ClientA", 2, "insert", "A", 1) // server: Aahllo ; clientA: AaAhello ; clientB: hello ; client C: Aahllo
	document = performOp(revlog, Op1, document)
	revlog[idx] = Op3
	idx++
	fmt.Println(document) // server: AaAhllo ; clientA: AaAhllo ; clientB: hello ; client C: Aahllo

	Op4 := getOp("ClientB", 1, "insert", "B", 5) // server: AaAhllo ; clientA: AaAhllo ; clientB: helloB ; client C: Aahllo
	document = performOp(revlog, Op1, document)
	revlog[idx] = Op4
	idx++
	fmt.Println(document) // server: AaAhlloB ; clientA: AaAhllo ; clientB: helloB ; client C: Aahllo

	Op5 := getOp("ClientC", 2, "delete", "l", 3) // server: AaAhlloB ; clientA: AaAhllo ; clientB: helloB ; client C: Aahlo
	document = performOp(revlog, Op1, document)
	revlog[idx] = Op5
	idx++
	fmt.Println(document) // server: AaAhloB ; clientA: AaAhllo ; clientB: helloB ; client C: AaAhloB

	Op6 := getOp("ClientB", 2, "delete", "B", 6) // server: AaAhloB ; clientA: AaAhllo ; clientB: hello ; client C: AaAhloB
	document = performOp(revlog, Op1, document)
	revlog[idx] = Op6
	idx++
	fmt.Println(document) // server: AaAhlo ; clientA: AaAhllo ; clientB: AaAhlo ; client C: AaAhloB

}

func main() {
	normalTest_L()
}
