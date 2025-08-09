package main

/*
#include <stdint.h>
*/
import "C"

import (
	"encoding/json"
	"funchooooza-ossh/loggo/core"
	"funchooooza-ossh/loggo/core/formatter"
	"funchooooza-ossh/loggo/core/writer"
	"sync"
	"time"
	"unsafe"
)

var (
	loggerStore              = map[uintptr]*core.Logger{}
	routeStore               = map[uintptr]*core.RouteProcessor{}
	formatStyleStore         = map[uintptr]*core.FormatStyle{}
	formatterStore           = map[uintptr]core.FormatProcessor{}
	writerStore              = map[uintptr]core.WriteProcessor{}
	compressorStore          = map[uintptr]core.WriteProcessor{}
	currentID        uintptr = 1
	storeMu          sync.Mutex
)

func makeID() uintptr {
	storeMu.Lock()
	defer storeMu.Unlock()
	id := currentID
	currentID++
	return id
}

//export NewLoggerWithRoutes
func NewLoggerWithRoutes(routeIDs *C.uintptr_t, count C.int) C.uintptr_t {
	routes := make([]*core.RouteProcessor, 0, int(count))

	// конвертация C-массива → Go-слайс
	slice := unsafe.Slice(routeIDs, count)

	for i := 0; i < int(count); i++ {
		r := routeStore[uintptr(slice[i])]
		if r != nil {
			routes = append(routes, r)
		}
	}

	logger := core.NewLogger(routes...)
	id := makeID()
	loggerStore[id] = logger
	return C.uintptr_t(id)
}

//export NewRouteProcessor
func NewRouteProcessor(formatterID, writerID C.uintptr_t, level C.uintptr_t) C.uintptr_t {
	formatter := formatterStore[uintptr(formatterID)]
	writer := writerStore[uintptr(writerID)]

	route := core.NewRouteProcessor(formatter, writer, core.LogLevel(level))
	id := makeID()
	routeStore[id] = route
	return C.uintptr_t(id)
}

//export NewStdoutWriter
func NewStdoutWriter() C.uintptr_t {
	writer := &writer.StdoutWriter{}
	id := makeID()
	writerStore[id] = writer
	return C.uintptr_t(id)
}

//export NewFileWriter
func NewFileWriter(path *C.char, maxSizeMB C.long, maxBackups C.int, interval *C.char, compress *C.char) C.uintptr_t {
	goPath := C.GoString(path)
	goInterval := writer.RotateInterval(C.GoString(interval))

	var goCompress *writer.Compress
	if compress != nil {
		c := writer.Compress(C.GoString(compress))
		goCompress = &c
	}

	writer, err := writer.NewFileWriter(
		goPath,
		int64(maxSizeMB), // ← безопасно привести C.long → int64
		int(maxBackups),
		goInterval,
		goCompress,
	)
	if err != nil {
		return 0
	}

	id := makeID()
	writerStore[id] = writer
	return C.uintptr_t(id)
}

//export NewTextFormatter
func NewTextFormatter(styleID C.uintptr_t, maxDepth C.int) C.uintptr_t {
	var style *core.FormatStyle
	if s, ok := formatStyleStore[uintptr(styleID)]; ok {
		style = s
	}
	depth := int(maxDepth)
	formatter := formatter.NewTextFormatter(style, &depth)
	id := makeID()
	formatterStore[id] = formatter
	return C.uintptr_t(id)
}

//export NewJsonFormatter
func NewJsonFormatter(styleId C.uintptr_t, maxDepth C.int) C.uintptr_t {
	var style *core.FormatStyle
	if s, ok := formatStyleStore[uintptr(styleId)]; ok {
		style = s
	}
	depth := int(maxDepth)
	formatter := formatter.NewJsonFormatter(style, &depth)
	id := makeID()
	formatterStore[id] = formatter
	return C.uintptr_t(id)
}

//export NewFormatStyle
func NewFormatStyle(colorKeys, colorValues, colorLevel C.uintptr_t, keyColor, valueColor, reset *C.char) C.uintptr_t {
	style := &core.FormatStyle{
		ColorKeys:   colorKeys != 0,
		ColorValues: colorValues != 0,
		ColorLevel:  colorLevel != 0,
		KeyColor:    C.GoString(keyColor),
		ValueColor:  C.GoString(valueColor),
		Reset:       C.GoString(reset),
	}
	id := makeID()
	formatStyleStore[id] = style
	return C.uintptr_t(id)
}

//export NewLoggerWithSingleRoute
func NewLoggerWithSingleRoute(routeID C.uintptr_t) C.uintptr_t {
	storeMu.Lock()
	defer storeMu.Unlock()

	route := routeStore[uintptr(routeID)]
	logger := core.NewLogger(route)
	id := makeID()
	loggerStore[id] = logger
	return C.uintptr_t(id)
}

func LogToRoute(routeId C.uintptr_t, level core.LogLevel, msg *C.char, fieldsJSON *C.char) {
	storeMu.Lock()
	route := routeStore[uintptr(routeId)]
	storeMu.Unlock()

	if !route.ShouldLog(level) {
		return
	}

	goMsg := C.GoString(msg)
	jsonStr := C.GoString(fieldsJSON)

	var fields map[string]interface{}
	_ = json.Unmarshal([]byte(jsonStr), &fields)
	if route == nil {
		return
	}

	record := core.LogRecord{
		Level:     level,
		Timestamp: time.Now(),
		Message:   goMsg,
		Fields:    fields,
	}
	route.Enqueue(record)

}

//export Logger_TraceToRoute
func Logger_TraceToRoute(routeId C.uintptr_t, msg *C.char, fields *C.char) {
	LogToRoute(routeId, core.Trace, msg, fields)
}

//export Logger_DebugToRoute
func Logger_DebugToRoute(routeId C.uintptr_t, msg *C.char, fields *C.char) {
	LogToRoute(routeId, core.Debug, msg, fields)
}

//export Logger_InfoToRoute
func Logger_InfoToRoute(routeId C.uintptr_t, msg *C.char, fields *C.char) {
	LogToRoute(routeId, core.Info, msg, fields)
}

//export Logger_WarningToRoute
func Logger_WarningToRoute(routeId C.uintptr_t, msg *C.char, fields *C.char) {
	LogToRoute(routeId, core.Warning, msg, fields)
}

//export Logger_ErrorToRoute
func Logger_ErrorToRoute(routeId C.uintptr_t, msg *C.char, fields *C.char) {
	LogToRoute(routeId, core.Error, msg, fields)
}

//export Logger_ExceptionToRoute
func Logger_ExceptionToRoute(routeId C.uintptr_t, msg *C.char, fields *C.char) {
	LogToRoute(routeId, core.Exception, msg, fields)
}

//export FreeLogger
func FreeLogger(loggerID C.uintptr_t) {
	storeMu.Lock()
	defer storeMu.Unlock()
	delete(loggerStore, uintptr(loggerID))
}

//export Logger_Close
func Logger_Close(loggerID C.uintptr_t) {
	storeMu.Lock()
	logger := loggerStore[uintptr(loggerID)]
	storeMu.Unlock()

	logger.Close()
}

func main() {}
