package util

import (
	"bufio"
	"fmt"
	"os"
	"sync/atomic"

	logutil "github.com/boz/go-logutil"
	lrlogutil "github.com/boz/go-logutil/logrus"
	"github.com/sirupsen/logrus"
)

const (
	cmpFieldName = "cmp"
)

var (
	currentID int32 = 0
)

type Env interface {
	Log() logrus.FieldLogger
	ForComponent(name string) Env
	WithFields(kvs ...string) Env
	WithID() Env

	LogErr(err error, fmt string, args ...interface{})

	FDebugf(fmt string, args ...interface{})
	Flush()

	Logutil() logutil.Log
}

func NewEnv(out *os.File, level string) (Env, error) {

	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return nil, err
	}

	logrus.StandardLogger().SetLevel(lvl)
	logrus.StandardLogger().Out = out

	log := logrus.New()
	log.Level = lvl
	log.Out = out

	return &env{log, out}, nil
}

func TestEnv() Env {
	log := logrus.New()
	log.Level = logrus.DebugLevel
	log.Out = os.Stderr
	return &env{log, os.Stderr}
}

type env struct {
	log logrus.FieldLogger
	out *os.File
}

func (e *env) Log() logrus.FieldLogger {
	return e.log
}

func (e *env) ForComponent(name string) Env {
	return e.WithFields(cmpFieldName, name)
}

func (e *env) WithFields(kvs ...string) Env {
	log := e.log
	for i := 1; i < len(kvs); i += 2 {
		log = log.WithField(kvs[i-1], kvs[i])
	}
	return &env{log, e.out}
}

func (e *env) WithID() Env {
	id := atomic.AddInt32(&currentID, 1)
	return e.WithFields("sid", fmt.Sprint(id))
}

func (e *env) LogErr(err error, fmt string, args ...interface{}) {
	e.log.WithError(err).Errorf(fmt, args...)
}

func (e *env) FDebugf(fmt string, args ...interface{}) {
	e.Log().Debugf(fmt, args...)
	e.Flush()
}

func (e *env) Flush() {
	bufio.NewWriter(e.out).Flush()
}

func (e *env) Logutil() logutil.Log {
	return lrlogutil.New(e.log)
}
