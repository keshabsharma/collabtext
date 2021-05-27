package main

import (
	"fmt"
)

func normalTest_L() {
	/*
	Clients: clientA, clientB, clientC
	Document: doc, initial with "hello"
	Revision: start with 1
	*/
	revlog := make([]Operation, 10)
	var idx int32 = 0
	document := "hello"

	Op1 := getOp("ClientA", 1, "insert", "Aa", 0)		// server: hello ; clientA: Aahello ; clientB: hello ; client C: hello
	document = performOp(revlog, Op1, document)		
	revlog[idx]=Op1
	idx++
	fmt.Println(document)	// server: Aahello ; clientA: Aahello ; clientB: hello ; client C: hello

	Op2 := getOp("ClientC", 1, "delete", "e", 2)	// server: Aahello ; clientA: Aahello ; clientB: hello ; client C: hllo
	document = performOp(revlog, Op1, document)
	revlog[idx]=Op2
	idx++
	fmt.Println(document)	// server: Aahllo ; clientA: Aahello ; clientB: hello ; client C: Aahllo

	Op3 := getOp("ClientA", 2, "insert", "A", 1)	// server: Aahllo ; clientA: AaAhello ; clientB: hello ; client C: Aahllo
	document = performOp(revlog, Op1, document)
	revlog[idx]=Op3
	idx++
	fmt.Println(document)	// server: AaAhllo ; clientA: AaAhllo ; clientB: hello ; client C: Aahllo

	Op4 := getOp("ClientB", 1, "insert", "B", 5)	// server: AaAhllo ; clientA: AaAhllo ; clientB: helloB ; client C: Aahllo
	document = performOp(revlog, Op1, document)
	revlog[idx]=Op4
	idx++
	fmt.Println(document)	// server: AaAhlloB ; clientA: AaAhllo ; clientB: helloB ; client C: Aahllo

	Op5 := getOp("ClientC", 2, "delete", "l", 3)	// server: AaAhlloB ; clientA: AaAhllo ; clientB: helloB ; client C: Aahlo
	document = performOp(revlog, Op1, document)
	revlog[idx]=Op5
	idx++
	fmt.Println(document)	// server: AaAhloB ; clientA: AaAhllo ; clientB: helloB ; client C: AaAhloB

	Op6 := getOp("ClientB", 2, "delete", "B", 6)	// server: AaAhloB ; clientA: AaAhllo ; clientB: hello ; client C: AaAhloB
	document = performOp(revlog, Op1, document)
	revlog[idx]=Op6
	idx++
	fmt.Println(document)	// server: AaAhlo ; clientA: AaAhllo ; clientB: AaAhlo ; client C: AaAhloB

}

func main() {
	normalTest_L()

}
