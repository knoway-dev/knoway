package protoutils

import (
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func TypeURLOrDie(obj proto.Message) string {
	a, err := anypb.New(obj)
	if err != nil {
		panic(err)
	}

	return a.GetTypeUrl()
}

func FromAny[T proto.Message](a *anypb.Any, prototype T) (T, error) {
	newObj, _ := reflect.New(reflect.TypeOf(prototype).Elem()).Interface().(T)
	if err := a.UnmarshalTo(newObj); err != nil {
		return newObj, err
	}

	return newObj, nil
}
