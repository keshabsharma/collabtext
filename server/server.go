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

type GetTransforms struct {
	Document     string
	FromRevision uint64
}

type TransformsList struct {
	Document string
	T        []t.Operation
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

	*ret = *processedOt

	return nil
}

func (s *Server) ApplyTransformations(transforms *TransformsList, rev *uint64) error {
	if len(transforms.T) == 0 {
		return fmt.Errorf("no transformations")
	}
	document, ok := s.documents[transforms.Document]
	if !ok {
		return fmt.Errorf("invalid document")
	}
	err := document.ApplyTransformations(transforms.T)
	if err != nil {
		return err
	}
	*rev = document.revision
	return nil
}

func (s *Server) GetTransformations(request GetTransforms, transforms *TransformsList) error {
	document, ok := s.documents[request.Document]
	if !ok {
		return fmt.Errorf("invalid document")
	}
	transforms.Document = request.Document
	transforms.T = document.GetTransformations(request.FromRevision)
	return nil
}

func (s *Server) init() {
	s.documents = make(map[string]*Document)
	s.documents["doc1"] = newDocument()
}

func RunServer(addr string, ready chan bool) error {
	server := rpc.NewServer()
	s := new(Server)
	s.init()

	err := server.Register(s)
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
