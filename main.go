package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	urlparse "net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/testutil/promlint"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"go.uber.org/multierr"
)

const (
	successExitCode = 0
	failureExitCode = 1
	// Exit code 3 is used for "one or more lint issues detected".
	lintErrExitCode = 3

	lintOptionAll                = "all"
	lintOptionDuplicateRules     = "duplicate-rules"
	lintOptionNone               = "none"
	dirData                      = "tmp/"
	chunksize                int = 1024
)

var (
	checkLink        bool = false
	checkCardinality bool = true
	sb               strings.Builder
	flagCacheDir     string
	flagHost         string
	flagPort         string
	flagTplDir       string
)

type metricStat struct {
	Name        string
	Cardinality int
	Percentage  float64
}

func floatToPercentage(x float64) float64 {
	return math.Round(x * 100)
}

func bytesToHuman(size int64) string {
	return humanize.Bytes(uint64(size))
}

func readCacheFile(filename string) (byteCount int, buffer *bytes.Buffer) {

	var data *os.File
	var part []byte
	var err error
	var count int

	data, err = os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer data.Close()

	reader := bufio.NewReader(data)
	buffer = bytes.NewBuffer(make([]byte, 0))
	part = make([]byte, chunksize)

	for {
		if count, err = reader.Read(part); err != nil {
			break
		}
		buffer.Write(part[:count])
	}
	if err != io.EOF {
		log.Fatal("Error Reading ", filename, ": ", err)
	} else {
		err = nil
	}

	return buffer.Len(), buffer

}

// ----------------------------

func setupRouter() *gin.Engine {
	// Disable Console Color
	gin.DisableConsoleColor()
	r := gin.Default()
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))
	//r.Delims("{[{", "}]}")
	r.SetFuncMap(template.FuncMap{
		"floatToPercentage": floatToPercentage,
		"bytesToHuman":      bytesToHuman,
	})

	r.LoadHTMLFiles(filepath.Join(flagTplDir, "analyze.tpl"))

	r.GET("/status", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	r.GET("/analyze", func(c *gin.Context) {

		var err error

		url := c.Query("url")
		resp, err := getContents(url)

		u, err := urlparse.Parse(url)
		if err != nil {
			c.HTML(http.StatusBadRequest, "analyze.tpl", gin.H{
				"url":   url,
				"error": err,
			})
		}

		contentFileName := fmt.Sprintf("%x_%s.prom", md5.Sum([]byte(url)), u.Hostname)
		cacheFilePath := filepath.Join(flagCacheDir, contentFileName)

		err = os.MkdirAll(flagCacheDir, os.ModePerm)
		if err != nil {
			c.HTML(http.StatusBadRequest, "analyze.tpl", gin.H{
				"url":   url,
				"error": err,
			})
		}
		out, err := os.Create(cacheFilePath)
		if err != nil {
			c.HTML(http.StatusBadRequest, "analyze.tpl", gin.H{
				"url":   url,
				"error": err,
			})
		}
		defer out.Close()
		nBytes, err := io.Copy(out, resp.Body)
		if err != nil {
			c.HTML(http.StatusBadRequest, "analyze.tpl", gin.H{
				"url":   url,
				"error": err,
			})
		}

		_, buffer := readCacheFile(cacheFilePath)
		resp.Body = ioutil.NopCloser(buffer)

		problems, err := checkMetricsLint(resp.Body)
		if err != nil {
			if err := os.Remove(cacheFilePath); err != nil {
				err = multierr.Append(err, err)
				c.HTML(http.StatusOK, "analyze.tpl", gin.H{
					"url":   url,
					"error": err.Error(),
				})
			}
		}

		_, buffer = readCacheFile(cacheFilePath)
		resp.Body = ioutil.NopCloser(buffer)

		stats, total, err := checkExtended(resp.Body)

		c.HTML(http.StatusOK, "analyze.tpl", gin.H{
			"url":                  url,
			"transferSize":         nBytes,
			"totalLintingProblems": len(problems),
			"lintingProblems":      problems,
			"resultCardinality":    stats,
			"totalMetrics":         total,
		})

	})

	return r
}

func getContents(url string) (*http.Response, error) {

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("User-Agent", "Prometheus Web Metric Analyzer/0.1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	fmt.Println("Code:", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status error: %v", resp.StatusCode)
	}
	return resp, nil

}

func checkMetricsLint(resp io.Reader) ([]promlint.Problem, error) {

	l := promlint.New(resp)
	problems, err := l.Lint()
	if err != nil {
		return []promlint.Problem{}, err
	}

	return problems, nil

}

func checkExtended(r io.Reader) ([]metricStat, int, error) {
	p := expfmt.TextParser{}
	metricFamilies, err := p.TextToMetricFamilies(r)
	if err != nil {
		return nil, 0, fmt.Errorf("error while parsing text to metric families: %w", err)
	}

	var total int
	stats := make([]metricStat, 0, len(metricFamilies))
	for _, mf := range metricFamilies {
		var cardinality int
		switch mf.GetType() {
		case dto.MetricType_COUNTER, dto.MetricType_GAUGE, dto.MetricType_UNTYPED:
			cardinality = len(mf.Metric)
		case dto.MetricType_HISTOGRAM:
			// Histogram metrics includes sum, count, buckets.
			buckets := len(mf.Metric[0].Histogram.Bucket)
			cardinality = len(mf.Metric) * (2 + buckets)
		case dto.MetricType_SUMMARY:
			// Summary metrics includes sum, count, quantiles.
			quantiles := len(mf.Metric[0].Summary.Quantile)
			cardinality = len(mf.Metric) * (2 + quantiles)
		default:
			cardinality = len(mf.Metric)
		}
		stats = append(stats, metricStat{Name: mf.GetName(), Cardinality: cardinality})
		total += cardinality
	}

	for i := range stats {
		stats[i].Percentage = float64(stats[i].Cardinality) / float64(total)
	}

	sort.SliceStable(stats, func(i, j int) bool {
		return stats[i].Cardinality > stats[j].Cardinality
	})

	return stats, total, nil
}

func main() {

	flag.StringVar(&flagCacheDir, "cachedir", "/tmp/prom-analyze-metrics-ui", "The path to the cache directory")
	flag.StringVar(&flagTplDir, "tpldir", "tpls/", "The path to html templates")
	flag.StringVar(&flagHost, "host", "", "Host on which to serve the web app")
	flag.StringVar(&flagPort, "port", "8080", "Port on which to serve the web app")
	flag.Parse()

	r := setupRouter()
	r.Run(fmt.Sprintf("%s:%s", flagHost, flagPort))
}
