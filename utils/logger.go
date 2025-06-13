package utils

import (
	"log"
	"strconv"
)

func LogError(svcName, refNo, methodName string, err error, data ...string) {
	str := svcName
	if refNo != "" {
		str += " , referenceNo :: " + refNo
	}
	if methodName != "" {
		str += " , error " + methodName
	}
	if err != nil {
		str += " :: " + err.Error()
	}
	for i, s := range data {
		str += ",\n Data " + strconv.Itoa(i+1) + " ::" + s
	}
	log.Println(str)
}

func LogInfo(svcName, refNo, methodName string, data ...string) {
	str := svcName
	if refNo != "" {
		str += " , referenceNo :: " + refNo
	}
	if methodName != "" {
		str += " , " + methodName
	}
	for i, s := range data {
		str += ",\n Data " + strconv.Itoa(i+1) + " ::" + s
	}
	log.Println(str)
}
