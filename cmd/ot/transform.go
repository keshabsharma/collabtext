package main

import (
	. "collabtext/transform"
	"fmt"
)

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
		op1.Position--
		return op1
	} else {
		//fmt.Println("huh")
		op1.Position = 0
		op1.Str = ""
		op1.Op = "insert"
		//op1.Op = "id"
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
		if op.Position >= int32(len(document)) {
			return document + op.Str
		} else {
			return document[0:op.Position] + op.Str + document[op.Position:]
		}
	} else {
		if op.Position >= int32(len(document)) {
			return document[0 : op.Position-1]
		} else {
			return document[0:op.Position] + document[op.Position+1:]
		}
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

func performOp(revlog []Operation, op Operation, document string) (string, Operation) {
	refOps := getRefOp(revlog, op)
	for _, rOp := range refOps {
		op = transform(op, rOp)
	}
	document = applyOp(document, op)
	return document, op
}

func normalTest_L() {
	/*
		Clients: clientA, clientB, clientC
		Document: doc, initial with "hello"
		Revision: start with 1
	*/
	revlog := make([]Operation, 0, 0)
	var idx int32 = 0
	document := "hello"

	revlog = append(revlog, getOp("clienta", 0, "insert", "", 0))

	Op1 := getOp("ClientA", 1, "insert", "A", 0) // server: hello ; clientA: Ahello ; clientB: hello ; client C: hello
	document, op_t := performOp(revlog, Op1, document)
	//revlog[idx] = op_t
	revlog = append(revlog, op_t)
	idx++
	fmt.Println(document) // server: Ahello ; clientA: Ahello ; clientB: hello ; client C: hello

	Op2 := getOp("ClientC", 1, "delete", "e", 1) // server: Aahello ; clientA: Aahello ; clientB: hello ; client C: hllo
	document, op_t = performOp(revlog, Op2, document)
	//revlog[idx] = op_t
	revlog = append(revlog, op_t)
	idx++
	fmt.Println(document) // server: Ahllo ; clientA: Ahello ; clientB: hello ; client C: Aahllo

	Op3 := getOp("ClientA", 2, "insert", "B", 1) // server: Ahllo ; clientA: ABhello ; clientB: hello ; client C: Ahllo
	document, op_t = performOp(revlog, Op3, document)
	revlog = append(revlog, op_t)
	idx++
	fmt.Println(document) // server: ABhllo ; clientA: ABhllo ; clientB: hello ; client C: Ahllo

	Op4 := getOp("ClientB", 1, "insert", "C", 5) // server: ABhlloC ; clientA: ABhllo ; clientB: helloC ; client C: Ahllo
	document, op_t = performOp(revlog, Op4, document)
	revlog = append(revlog, op_t)
	idx++
	fmt.Println(document) // server: ABhlloC ; clientA: ABhllo ; clientB: helloC ; client C: Ahllo

	Op5 := getOp("ClientC", 2, "delete", "l", 3) // server: AaAhlloB ; clientA: AaAhllo ; clientB: helloB ; client C: Aahlo
	document, op_t = performOp(revlog, Op5, document)
	revlog = append(revlog, op_t)
	idx++
	fmt.Println(document) // server: AaAhloB ; clientA: AaAhllo ; clientB: helloB ; client C: AaAhloB

	Op6 := getOp("ClientB", 2, "delete", "B", 6) // server: AaAhloB ; clientA: AaAhllo ; clientB: hello ; client C: AaAhloB
	document, op_t = performOp(revlog, Op6, document)
	revlog = append(revlog, op_t)
	idx++
	fmt.Println(document) // server: AaAhlo ; clientA: AaAhllo ; clientB: AaAhlo ; client C: AaAhloB

}

func ddtest() {

	revlog := make([]Operation, 0, 0)
	var idx int32 = 0
	document := "a"

	revlog = append(revlog, getOp("clienta", 0, "insert", "", 0))

	Op1 := getOp("ClientA", 1, "delete", "A", 0)
	document, op_t := performOp(revlog, Op1, document)
	//revlog[idx] = op_t
	revlog = append(revlog, op_t)
	idx++
	fmt.Println(document)

	Op1 = getOp("ClientA", 2, "insert", "B", 0)
	document, op_t = performOp(revlog, Op1, document)
	//revlog[idx] = op_t
	revlog = append(revlog, op_t)
	idx++
	fmt.Println(document)

	Op2 := getOp("Client2", 1, "delete", "A", 0)
	document, op_t = performOp(revlog, Op2, document)
	//revlog[idx] = op_t
	revlog = append(revlog, op_t)
	idx++
	fmt.Println(document)

}

func main() {
	//normalTest_L()
	ddtest()
}
