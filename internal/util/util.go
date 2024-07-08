package util

import (
	"net"
	"reflect"
)

func GetIP() string {
	conn, error := net.Dial("udp", "8.8.8.8:80")
	if error != nil {
		return ""
	}

	defer conn.Close()
	addr := conn.LocalAddr().(*net.UDPAddr)

	return addr.IP.String()
}

func UpdateValue(newValue any, obj any) bool {
	var isupdate bool
	var ntk reflect.Kind
	var otk reflect.Kind

	nt := reflect.TypeOf(newValue)
	nv := reflect.ValueOf(newValue)
	ntk = nt.Kind()
	if ntk == reflect.Ptr {
		nt = nt.Elem()
		nv = nv.Elem()
	}
	if nt.Kind() != reflect.Struct {
		return isupdate
	}

	ot := reflect.TypeOf(obj)
	ov := reflect.ValueOf(obj)
	otk = ot.Kind()
	if otk != reflect.Ptr {
		return isupdate
	}

	if ntk == otk && reflect.DeepEqual(newValue, obj) {
		return isupdate
	}

	ot = ot.Elem()
	ov = ov.Elem()
	if ot.Kind() != reflect.Struct {
		return isupdate
	}

	vtFieldNum := nt.NumField()
	for i := 0; i < vtFieldNum; i++ {
		ovv := ov.FieldByName(nt.Field(i).Name)
		if ovv.IsValid() && ovv.CanSet() {
			nvv := nv.Field(i)
			if !nvv.IsZero() && nvv.Type() == ovv.Type() && !nvv.Equal(ovv) {
				ovv.Set(nvv)
				isupdate = true
			}
		}
	}
	return isupdate
}
