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

	humanize "github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/testutil/promlint"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
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
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.RedirectTrailingSlash = false // not default
	r.RedirectFixedPath = false
	r.HandleMethodNotAllowed = false
	r.ForwardedByClientIP = true

	// https://golang.org/src/net/url/url.go
	r.UseRawPath = false
	r.UnescapePathValues = true

	r.Use(JsonLogger())
	//r.Delims("{[{", "}]}")

	r.SetFuncMap(template.FuncMap{
		"floatToPercentage": floatToPercentage,
		"bytesToHuman":      bytesToHuman,
	})

	r.LoadHTMLFiles(
		filepath.Join(flagTplDir, "analyze.tpl"),
		filepath.Join(flagTplDir, "analyze_get.tpl"),
	)

	r.GET("/status", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "analyze_get.tpl", gin.H{})
	})

	r.POST("/analyze", func(c *gin.Context) {

		url := c.PostForm("url")

		resp, err := getContents(url)
		if err != nil {
			showError(c, url, err)
			return
		}

		u, err := urlparse.Parse(url)
		if err != nil {
			showError(c, url, err)
			return
		}

		contentFileName := fmt.Sprintf("%x_%s.prom", md5.Sum([]byte(url)), u.Hostname())
		cacheFilePath := filepath.Join(flagCacheDir, contentFileName)

		err = os.MkdirAll(flagCacheDir, os.ModePerm)
		if err != nil {
			showError(c, url, err)
			return
		}

		out, err := os.Create(cacheFilePath)
		defer out.Close()
		if err != nil {
			showError(c, url, err)
			return
		}

		defer out.Close()

		nBytes, err := io.Copy(out, resp.Body)
		if err != nil {
			showError(c, url, err)
			return
		}

		_, buffer := readCacheFile(cacheFilePath)
		resp.Body = ioutil.NopCloser(buffer)

		problems, err := checkMetricsLint(resp.Body)
		if err != nil {
			// Remove metrics cache file if present
			if fileExists(cacheFilePath) {
				if err := os.Remove(cacheFilePath); err != nil {
					err = fmt.Errorf("%w; %w", err, err)
				}
			}

		}
		if err != nil {
			showError(c, url, err)
			return
		}

		_, buffer = readCacheFile(cacheFilePath)
		resp.Body = ioutil.NopCloser(buffer)
		stats, total, err := checkExtended(resp.Body)

		err = multierr.Append(err, err)

		if err := os.Remove(cacheFilePath); err != nil {
			err = multierr.Append(err, err)
			c.HTML(http.StatusOK, "analyze.tpl", gin.H{
				"url":   url,
				"error": err.Error(),
			})
		}

		c.HTML(http.StatusOK, "analyze.tpl", gin.H{
			"url":                  url,
			"transferSize":         nBytes,
			"totalLintingProblems": len(problems),
			"lintingProblems":      problems,
			"resultCardinality":    stats,
			"totalMetrics":         total,
			"error":                err,
		})
		return

	})

	return r
}

func fileExists(filePath string) bool {
	if _, error := os.Stat(filePath); os.IsNotExist(error) {
		return false
	}
	return true
}

func showError(c *gin.Context, url string, err error) {
	c.HTML(http.StatusOK, "analyze.tpl", gin.H{
		"url":   url,
		"error": err,
	})
}

func getContents(url string) (*http.Response, error) {

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("User-Agent", "Prometheus Web Metric Verifier/0.1.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return resp, fmt.Errorf("Response error: %v", resp.Status)
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
