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

func FromAny[T proto.Message](a *anypb.Any) (T, error) {
	var obj T

	objType := reflect.TypeOf(obj).Elem() // 获取目标类型
	newObj, _ := reflect.New(objType).Interface().(T)

	err := a.UnmarshalTo(newObj)
	if err != nil {
		return obj, err
	}

	return newObj, nil
}
