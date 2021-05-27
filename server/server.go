package server

import (
	t "collabtext/transform"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

type Server struct {
	documents map[string]*Document
}

func (s *Server) ProcessTransformation(ot *t.Operation, ret *t.Operation) error {
	document, ok := s.documents[ot.Document]
	if !ok {
		return fmt.Errorf("invalid document")
	}

	processedOt, err := document.ProcessTransformation(ot)
	if err != nil {
		return err
	}

	ret = processedOt
	return nil
}

func (s *Server) ApplyTransformations(ots []*t.Operation, succ *bool) error {
	document, ok := s.documents[ots[0].Document]
	if !ok {
		return fmt.Errorf("invalid document")
	}
	err := document.ApplyTransformations(ots)
	if err != nil {
		return err
	}
	*succ = true
	return nil
}

func RunServer(addr string, ready chan bool) error {
	server := rpc.NewServer()
	err := server.Register(new(Server))
	if err != nil {
		ready <- false
		return err
	}

	listener, error := net.Listen("tcp", addr)
	if error != nil {
		ready <- false
		return error
	}
	ready <- true

	log.Println("serving server ", addr)

	return http.Serve(listener, server)
}
