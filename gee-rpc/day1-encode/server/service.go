package server

import (
	"geerpc/protocol"
	"go/ast"
	"log"
	"reflect"
)

type methodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
}

func (m *methodType) NewArg() interface{} {
	return newTypeInter(m.ArgType)
}

func (m *methodType) NewReply() interface{} {
	return newTypeInter(m.ReplyType)
}

func newTypeInter(t reflect.Type) interface{} {
	var v reflect.Value
	if t.Kind() == reflect.Ptr { // reply must be ptr
		v = reflect.New(t.Elem())
	} else {
		v = reflect.New(t)
	}
	return v.Interface()
}

type service struct {
	name   string
	rcvr   reflect.Value
	method map[string]*methodType
}

func newService(receiver interface{}) *service {
	service := new(service)
	service.method = make(map[string]*methodType)
	service.name = reflect.Indirect(reflect.ValueOf(receiver)).Type().Name()
	service.rcvr = reflect.ValueOf(receiver)

	_assert(ast.IsExported(service.name), "%service is not exported", service.name)
	rcvrType := reflect.TypeOf(receiver)
	for i := 0; i < rcvrType.NumMethod(); i++ {
		method := rcvrType.Method(i)
		mType := method.Type
		if mType.NumIn() != 3 || mType.NumOut() != 1 {
			continue
		}
		if mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}

		argType, replyType := mType.In(1), mType.In(2)
		if !isExportedOrBuiltinType(argType) || !isExportedOrBuiltinType(replyType) {
			continue
		}

		service.method[method.Name] = &methodType{
			method:    method,
			ArgType:   argType,
			ReplyType: replyType,
		}
		log.Printf("Register %s.%s\n", service.name, method.Name)
	}

	return service
}

func (s *service) call(methodName string, reqMsg *protocol.Message) (resp *protocol.Message) {
	mType := s.method[methodName]
	resp = reqMsg.Clone()

	arg, reply := mType.NewArg(), mType.NewReply()
	if err := reqMsg.GetPayload(arg); err != nil {
		return resp.HandleError(protocol.ExecError, err)
	}

	f := mType.method.Func
	returnValues := f.Call([]reflect.Value{s.rcvr, reflect.ValueOf(arg).Elem(), reflect.ValueOf(reply)})

	if errInter := returnValues[0].Interface(); errInter != nil {
		return resp.HandleError(protocol.ExecError, errInter.(error))
	}

	if err := resp.SetPayload(reply); err != nil {
		return resp.HandleError(protocol.ExecError, err)
	}

	return resp
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}
