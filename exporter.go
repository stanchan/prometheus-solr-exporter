package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const (
	mbeansPath      = "/admin/mbeans?stats=true&wt=json&cat=CORE&cat=QUERY&cat=UPDATE&cat=CACHE"
	adminCoresPath  = "/admin/cores?action=STATUS&wt=json"
	metricsCorePath = "/admin/metrics?group=core&prefix=QUERY,UPDATE&wt=json"
)

var (
	gaugeAdminMetrics = map[string]string{
		"num_docs":      "num_docs",
		"size_in_bytes": "size_in_bytes",
		"deleted_docs":  "deleted_docs",
		"max_docs":      "max_docs",
	}
	gaugeCoreMetrics = map[string]string{
		"num_docs":     "num_docs",
		"deleted_docs": "deleted_docs",
		"max_docs":     "max_docs",
	}
	gaugeQueryMetrics = map[string]string{
		"15min_rate_reqs_per_second": "15min_rate_reqs_per_second",
		"5min_rate_reqs_per_second":  "5min_rate_reqs_per_second",
		"75th_pc_request_time":       "75th_pc_request_time",
		"95th_pc_request_time":       "95th_pc_request_time",
		"99th_pc_request_time":       "99th_pc_request_time",
		"999th_pc_request_time":      "999th_pc_request_time",
		"avg_requests_per_second":    "avg_requests_per_second",
		"avg_time_per_request":       "avg_time_per_request",
		"median_time_per_request":    "median_time_per_request",
		"requests":                   "requests",
		"errors":                     "errors",
		"client_errors":              "client_errors",
		"server_errors":              "server_errors",
		"handler_start":              "handler_start",
		"timeouts":                   "timeouts",
		"total_time":                 "total_time",
	}
	gaugeUpdateMetrics = map[string]string{
		"15min_rate_updates_per_second": "15min_rate_updates_per_second",
		"5min_rate_updates_per_second":  "5min_rate_updates_per_second",
		"75th_pc_update_time":           "75th_pc_update_time",
		"95th_pc_update_time":           "95th_pc_update_time",
		"99th_pc_update_time":           "99th_pc_update_time",
		"999th_pc_update_time":          "999th_pc_update_time",
		"avg_updates_per_second":        "avg_updates_per_second",
		"avg_time_per_update":           "avg_time_per_update",
		"median_time_per_update":        "median_time_per_update",
		"requests":                      "requests",
		"adds":                          "adds",
		"autocommit_max_docs":           "autocommit_max_docs",
		"autocommit_max_time":           "autocommit_max_time",
		"autocommits":                   "autocommits",
		"commits":                       "commits",
		"cumulative_adds":               "cumulative_adds",
		"cumulative_deletes_by_id":      "cumulative_deletes_by_id",
		"cumulative_deletes_by_query":   "cumulative_deletes_by_query",
		"cumulative_errors":             "cumulative_errors",
		"deletes_by_id":                 "deletes_by_id",
		"deletes_by_query":              "deletes_by_query",
		"docs_pending":                  "docs_pending",
		"errors":                        "errors",
		"expunge_deletes":               "expunge_deletes",
		"merges":                        "merges",
		"optimizes":                     "optimizes",
		"rollbacks":                     "rollbacks",
		"soft_autocommits":              "soft_autocommits",
	}
	gaugeCacheMetrics = map[string]string{
		"cumulative_evictions": "cumulative_evictions",
		"cumulative_hitratio":  "cumulative_hitratio",
		"cumulative_hits":      "cumulative_hits",
		"cumulative_inserts":   "cumulative_inserts",
		"cumulative_lookups":   "cumulative_lookups",
		"evictions":            "evictions",
		"hitratio":             "hitratio",
		"hits":                 "hits",
		"inserts":              "inserts",
		"lookups":              "lookups",
		"size":                 "size",
		"warmup_time":          "warmup_time",
	}
)

// Return list of cores from solr server
func getCoresFromStatus(adminCoresStatus *AdminCoresStatus) []string {
	serverCores := []string{}
	for coreName := range adminCoresStatus.Status {
		serverCores = append(serverCores, coreName)
	}
	return serverCores
}

// Exporter collects Solr stats from the given server and exports
// them using the prometheus metrics package.
type Exporter struct {
	mBeansURL    string
	adminCoreURL string
	metricsURL   string
	mutex        sync.RWMutex

	up prometheus.Gauge

	gaugeAdmin  map[string]*prometheus.GaugeVec
	gaugeCore   map[string]*prometheus.GaugeVec
	gaugeQuery  map[string]*prometheus.GaugeVec
	gaugeUpdate map[string]*prometheus.GaugeVec
	gaugeCache  map[string]*prometheus.GaugeVec

	client http.Client
}

// NewExporter returns an initialized Exporter.
func NewExporter(solrBaseURL string, timeout time.Duration, solrExcludedCore string, client http.Client) *Exporter {
	gaugeAdmin := make(map[string]*prometheus.GaugeVec, len(gaugeAdminMetrics))
	gaugeCore := make(map[string]*prometheus.GaugeVec, len(gaugeCoreMetrics))
	gaugeQuery := make(map[string]*prometheus.GaugeVec, len(gaugeQueryMetrics))
	gaugeUpdate := make(map[string]*prometheus.GaugeVec, len(gaugeUpdateMetrics))
	gaugeCache := make(map[string]*prometheus.GaugeVec, len(gaugeCacheMetrics))

	for name, help := range gaugeAdminMetrics {
		gaugeAdmin[name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace + "_admin",
			Name:      name,
			Help:      help,
		}, []string{"core"})
	}

	for name, help := range gaugeCoreMetrics {
		gaugeCore[name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace + "_core",
			Name:      name,
			Help:      help,
		}, []string{"core", "handler", "class"})
	}

	for name, help := range gaugeQueryMetrics {
		gaugeQuery[name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace + "_queryhandler",
			Name:      name,
			Help:      help,
		}, []string{"core", "handler", "class"})
	}

	for name, help := range gaugeUpdateMetrics {
		gaugeUpdate[name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace + "_updatehandler",
			Name:      name,
			Help:      help,
		}, []string{"core", "handler", "class"})
	}
	for name, help := range gaugeCacheMetrics {
		gaugeCache[name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace + "_cache",
			Name:      name,
			Help:      help,
		}, []string{"core", "handler", "class"})
	}

	mBeansURL := fmt.Sprintf("%s%s%s", solrBaseURL, "%s", mbeansPath)
	adminCoreURL := fmt.Sprintf("%s%s", solrBaseURL, adminCoresPath)
	metricsURL := fmt.Sprintf("%s%s", solrBaseURL, metricsCorePath)

	// Init our exporter.
	return &Exporter{
		mBeansURL:    mBeansURL,
		adminCoreURL: adminCoreURL,
		metricsURL:   metricsURL,

		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Was the Solr instance query successful?",
		}),

		gaugeAdmin:  gaugeAdmin,
		gaugeCore:   gaugeCore,
		gaugeQuery:  gaugeQuery,
		gaugeUpdate: gaugeUpdate,
		gaugeCache:  gaugeCache,

		client: client,
	}
}

// Describe describes all the metrics ever exported by the solr
// exporter. It implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up.Desc()

	for _, vec := range e.gaugeAdmin {
		vec.Describe(ch)
	}
	for _, vec := range e.gaugeCore {
		vec.Describe(ch)
	}
	for _, vec := range e.gaugeQuery {
		vec.Describe(ch)
	}
	for _, vec := range e.gaugeUpdate {
		vec.Describe(ch)
	}
	for _, vec := range e.gaugeCache {
		vec.Describe(ch)
	}
}

// Collect fetches the stats from configured solr location and
// delivers them as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()

	// Reset metrics.
	for _, vec := range e.gaugeAdmin {
		vec.Reset()
	}
	for _, vec := range e.gaugeCore {
		vec.Reset()
	}
	for _, vec := range e.gaugeQuery {
		vec.Reset()
	}
	for _, vec := range e.gaugeUpdate {
		vec.Reset()
	}
	for _, vec := range e.gaugeCache {
		vec.Reset()
	}

	e.up.Set(0)
	defer func() { ch <- e.up }()

	resp, err := e.client.Get(e.adminCoreURL)
	if err != nil {
		log.Errorf("Error while querying Solr for admin stats: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Failed to read admin stats response body: %v", err)
		return
	}

	solrExcludedCoreString := *solrExcludedCore
	var regexExludedCore = regexp.MustCompile(solrExcludedCoreString)

	adminCoresStatus := &AdminCoresStatus{}
	err = json.Unmarshal(body, adminCoresStatus)
	if err != nil {
		log.Errorf("Failed to unmarshal solr admin JSON into struct: %v", err)
		return
	}

	for core, metrics := range adminCoresStatus.Status {
		if solrExcludedCoreString != "" && regexExludedCore.MatchString(core) {
			continue
		}
		e.gaugeAdmin["num_docs"].WithLabelValues(core).Set(float64(metrics.Index.NumDocs))
		e.gaugeAdmin["size_in_bytes"].WithLabelValues(core).Set(float64(metrics.Index.SizeInBytes))
		e.gaugeAdmin["deleted_docs"].WithLabelValues(core).Set(float64(metrics.Index.DeletedDocs))
		e.gaugeAdmin["max_docs"].WithLabelValues(core).Set(float64(metrics.Index.MaxDoc))
	}

	cores := getCoresFromStatus(adminCoresStatus)

	metricsResp, err := e.client.Get(e.metricsURL)
	if err != nil {
		log.Errorf("Error while querying Solr for metrics: %v", err)
		return
	}
	defer metricsResp.Body.Close()

	if metricsResp.StatusCode != http.StatusOK {
		log.Errorf("solr: API responded with status-code %d, expected %d, url %s",
			metricsResp.StatusCode, http.StatusOK, e.metricsURL)
		return
	}

	metricsBody, err := ioutil.ReadAll(metricsResp.Body)
	if err != nil {
		log.Errorf("Failed to read Solr metrics response body: %v", err)
		return
	}

	for _, coreName := range cores {
		if solrExcludedCoreString != "" && regexExludedCore.MatchString(coreName) {
			continue
		}
		mBeansURL := fmt.Sprintf(e.mBeansURL, "/"+coreName)
		mBeansURLResp, err := e.client.Get(mBeansURL)
		if err != nil {
			log.Errorf("Error while querying Solr for mbeans stats: %v", err)
			return
		}
		defer mBeansURLResp.Body.Close()

		if mBeansURLResp.StatusCode != http.StatusOK {
			log.Errorf("solr: API responded with status-code %d, expected %d, url %s",
				mBeansURLResp.StatusCode, http.StatusOK, mBeansURL)
			return
		}

		mBeansBody, err := ioutil.ReadAll(mBeansURLResp.Body)
		if err != nil {
			log.Errorf("Failed to read Solr mbeans response body: %v", err)
			return
		}

		errors := processMbeans(e, coreName, mBeansBody)
		for _, err := range errors {
			log.Error(err)
		}

		errors = processQueryMetrics(e, coreName, metricsBody)
		for _, err := range errors {
			log.Error(err)
		}

		errors = processUpdateMetrics(e, coreName, metricsBody)
		for _, err := range errors {
			log.Error(err)
		}
	}

	// Report metrics.
	for _, vec := range e.gaugeAdmin {
		vec.Collect(ch)
	}
	for _, vec := range e.gaugeCore {
		vec.Collect(ch)
	}
	for _, vec := range e.gaugeQuery {
		vec.Collect(ch)
	}
	for _, vec := range e.gaugeUpdate {
		vec.Collect(ch)
	}
	for _, vec := range e.gaugeCache {
		vec.Collect(ch)
	}

	// Successfully processed stats.
	e.up.Set(1)
}
