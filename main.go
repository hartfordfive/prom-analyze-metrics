package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/testutil/promlint"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"

	"text/tabwriter"
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
	name        string
	cardinality int
	percentage  float64
}

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	r.LoadHTMLFiles("./tpls/analyze.tpl")

	r.GET("/status", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// Get user value
	r.GET("/analyze", func(c *gin.Context) {

		outputLinting := ""
		outputCardinality := ""

		url := c.Query("url")

		fmt.Println("URL: ", url)

		if checkLink {
			_, err := checkMetricsLint(url)
			if err != nil {
				outputLinting = fmt.Sprintf("error while linting:", err)

			}
		}
		if checkCardinality {
			_, outputCardinality = checkMetricsExtended(url)
		}

		c.HTML(http.StatusOK, "analyze.tpl", gin.H{
			"resultLinting":     outputLinting,
			"resultCardinality": outputCardinality,
		})

	})

	return r
}

func main() {

	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}

func getContents(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status error: %v", resp.StatusCode)
	}
	return resp.Body, nil

}

func checkMetricsLint(url string) (int, error) {

	resp, err := getContents(url)

	l := promlint.New(resp)
	problems, err := l.Lint()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error while linting:", err)
		return failureExitCode, err
	}

	for _, p := range problems {
		fmt.Fprintln(os.Stderr, p.Metric, p.Text)
	}

	if len(problems) > 0 {
		return lintErrExitCode, nil
	}

	return successExitCode, nil
}

func checkMetricsExtended(url string) (int, string) {
	var buf bytes.Buffer
	stats, total, err := checkExtended(&buf)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return failureExitCode, ""
	}
	w := tabwriter.NewWriter(os.Stdout, 4, 4, 4, ' ', tabwriter.TabIndent)
	//fmt.Sprintf(w, "Metric\tCardinality\tPercentage\t\n")
	sb.WriteString("Metric\tCardinality\tPercentage\t\n")
	for _, stat := range stats {
		//fmt.Sprintf(w, "%s\t%d\t%.2f%%\t\n", stat.name, stat.cardinality, stat.percentage*100)
		sb.WriteString(fmt.Sprintf("%s\t%d\t%.2f%%\t\n", stat.name, stat.cardinality, stat.percentage*100))
	}
	//fmt.Sprintf(w, "Total\t%d\t%.f%%\t\n", total, 100.)
	sb.WriteString(fmt.Sprintf("Total\t%d\t%.f%%\t\n", total, 100.))
	w.Flush()

	return successExitCode, sb.String()
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
		stats = append(stats, metricStat{name: mf.GetName(), cardinality: cardinality})
		total += cardinality
	}

	for i := range stats {
		stats[i].percentage = float64(stats[i].cardinality) / float64(total)
	}

	sort.SliceStable(stats, func(i, j int) bool {
		return stats[i].cardinality > stats[j].cardinality
	})

	return stats, total, nil
}
