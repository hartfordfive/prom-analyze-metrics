// (C)2020 Tuomo Kuure
// JSON logger middleware for gin-gonic

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type LogItems struct {
	ISOTime        time.Time
	UnixTime       int64
	IP             string
	Method         string
	Host           string
	User           string
	Path           string
	Query          string
	Protocol       string
	ContentType    string
	ContentLength  int64
	ResponseStatus int
	ResponseSize   int
	Headers        http.Header
	TLSData        TLSData

	RequestProcessingTime int64
	LogProcessingTime     int64
}

type TLSData struct {
	TLSVersion     uint16
	TLSCipherUsed  uint16
	TLSMutualProto bool
}

var FormatJSON = func(log LogItems) string {
	logline, _ := json.Marshal(log)
	return fmt.Sprintf("%s\n", logline)
}

//func JsonLogger(filename string, w_stdout bool) gin.HandlerFunc {
func JsonLogger() gin.HandlerFunc {

	// TODO: When is the right time to close the log file?
	// logfile, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// // Write both to the file and stdout if desired
	// // TODO: Pretty-print to stdout option?
	// var out = io.Writer(logfile)
	// if w_stdout {
	// 	out = io.MultiWriter(logfile, os.Stdout)
	// }

	out := io.MultiWriter(os.Stdout)

	return func(c *gin.Context) {

		// All time values are in nanoseconds
		start := time.Now()
		c.Next() // Request is processed here
		stop := time.Now().UnixNano()

		log := LogItems{
			ISOTime: start,
			//UnixTime:       start.UnixNano(),
			IP:             c.ClientIP(),
			Method:         c.Request.Method,
			Host:           c.Request.Host,
			User:           c.Request.URL.User.Username(),
			Path:           c.Request.URL.EscapedPath(),
			Query:          c.Request.URL.RawQuery,
			Protocol:       c.Request.Proto,
			ContentType:    c.ContentType(),
			ContentLength:  c.Request.ContentLength,
			ResponseStatus: c.Writer.Status(),
			ResponseSize:   c.Writer.Size(),
			// Headers are now placed arrays, eg. "Dnt":["1"]
			// Header type is not a struct, but is the format a problem?
			//Headers: c.Request.Header,
		}

		if c.Request.TLS != nil {
			// https://golang.org/pkg/crypto/tls/#pkg-constants
			log.TLSData = TLSData{
				TLSVersion:     c.Request.TLS.Version,
				TLSCipherUsed:  c.Request.TLS.CipherSuite,
				TLSMutualProto: c.Request.TLS.NegotiatedProtocolIsMutual,
			}
		}

		log.RequestProcessingTime = stop - log.UnixTime

		// Measure the time it took to process the log
		// TODO: this should be the very last operation
		log.LogProcessingTime = time.Now().UnixNano() - stop

		fmt.Fprint(out, FormatJSON(log))
	}
}
