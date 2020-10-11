package cutil

// #include "debug_util.hxx"
import "C"

import (
  "sync"
)

var mutex sync.Mutex

func GetThreadID() string {
  mutex.Lock()
  threadId := C.GoString(C.getThreadId())
  defer mutex.Unlock()
  return threadId
}

func GetLogTm() string {
  mutex.Lock()
  logTm := C.GoString(C.getLogTm())
  defer mutex.Unlock()
  return logTm
}
