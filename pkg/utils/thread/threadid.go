package thread

/*
#cgo CFLAGS: -I.

#include "debug_util.hxx"

#include <stdlib.h>
#include <stdio.h>

void myPrintFunction2() {
        printf("Hello from inline C\n");
}
*/
import "C"
import (
	"runtime"
	"time"

	"github.com/go-logr/logr"
)

func WrapLog(realLogger logr.Logger, l int) logr.Logger {
	wrapper := &ThreadLoggerImpl{
		level:  l,
		logger: realLogger,
	}
	return wrapper
}

type ThreadLoggerImpl struct {
	level  int
	logger logr.Logger
}

func (impl *ThreadLoggerImpl) Enabled() bool {
	return impl.logger.Enabled()
}

func (impl *ThreadLoggerImpl) Info(msg string, keysAndValues ...interface{}) {
	//important! go may switch threads before calling C lib!! wtf!
	//Note this is not enough, you need to add those Lock/UnlockOSThread methods
	//in the calling go routines !!
	// one more thing: build with CGO_ENABLED=1
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	//impl.logger.Info("["+GetThreadID()+"]("+GetCurrentTime()+"): "+msg, keysAndValues...)
	impl.logger.Info("["+GetThreadID()+"] "+msg, keysAndValues...)
}

func (impl *ThreadLoggerImpl) Error(err error, msg string, keysAndValues ...interface{}) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	impl.logger.Error(err, "["+GetThreadID()+"] "+msg, keysAndValues...)
}

func (impl *ThreadLoggerImpl) V(level int) logr.Logger {
	return WrapLog(impl.logger.V(level), level)
}

func (impl *ThreadLoggerImpl) WithValues(keysAndValues ...interface{}) logr.Logger {
	return WrapLog(impl.logger.WithValues(keysAndValues...), impl.level)
}

func (impl *ThreadLoggerImpl) WithName(name string) logr.Logger {
	return WrapLog(impl.logger.WithName(name), impl.level)
}

func GetThreadID() string {
	threadId := C.GoString(C.getThreadId())
	return threadId
}

func GetCurrentTime() string {
	return time.Now().Format("01-02-2006 15:04:05.000000")
}
