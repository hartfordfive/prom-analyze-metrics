package main

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

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

	lintOptionAll            = "all"
	lintOptionDuplicateRules = "duplicate-rules"
	lintOptionNone           = "none"
)

var (
	//url              string = "http://localhost:9100/metrics"
	checkLink        bool = false
	checkCardinality bool = true
	sb               strings.Builder
)

type metricStat struct {
	Name        string
	Cardinality int
	Percentage  float64
}

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	r.LoadHTMLFiles("./tpls/analyze.tpl")

	r.GET("/status", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	r.GET("/analyze", func(c *gin.Context) {

		//outputLinting := ""

		url := c.Query("url")

		//fmt.Println("URL: ", url)

		problems, err := checkMetricsLint(url)
		if err != nil {
			c.HTML(http.StatusOK, "analyze.tpl", gin.H{
				"resultLinting":     "",
				"resultCardinality": "",
				"totalMetrics":      "",
				"error":             err.Error(),
			})
		}

		resp, err := getContents(url)
		if err != nil {
			c.HTML(http.StatusOK, "analyze.tpl", gin.H{
				"resultLinting":     "",
				"resultCardinality": "",
				"totalMetrics":      "",
				"error":             err.Error(),
			})
		}

		stats, total, err := checkExtended(resp)

		c.HTML(http.StatusOK, "analyze.tpl", gin.H{
			"totalLintingProblems": len(problems),
			"lintingProblems":      problems,
			"resultCardinality":    stats,
			"totalMetrics":         total,
		})

	})

	return r
}

func getContents(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GET error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status error: %v", resp.StatusCode)
	}
	return resp.Body, nil

}

func checkMetricsLint(url string) ([]promlint.Problem, error) {

	resp, err := getContents(url)
	if err != nil {
		return []promlint.Problem{}, err
	}

	l := promlint.New(resp)
	problems, err := l.Lint()
	if err != nil {
		//fmt.Fprintln(os.Stderr, "error while linting:", err)
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

	r := setupRouter()
	r.Run(":8080")
}
