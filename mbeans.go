package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func findMBeansData(mBeansData []json.RawMessage, query string) json.RawMessage {
	var decoded string
	for i := 0; i < len(mBeansData); i++ {
		err := json.Unmarshal(mBeansData[i], &decoded)
		if err == nil {
			if decoded == query || decoded == query+"HANDLER" {
				return mBeansData[i+1]
			}
		}
	}

	return nil
}

func processMbeans(e *Exporter, coreName string, data io.Reader) []error {
	mBeansData := &MBeansData{}
	errors := []error{}
	if err := json.NewDecoder(data).Decode(mBeansData); err != nil {
		errors = append(errors, fmt.Errorf("Failed to unmarshal mbeansdata JSON into struct: %v", err))
		return errors
	}

	var coreMetrics map[string]Core
	coreData := findMBeansData(mBeansData.SolrMbeans, "CORE")
	b := bytes.Replace(coreData, []byte(":\"NaN\""), []byte(":0.0"), -1)
	b = bytes.Replace(b, []byte("SEARCHER.searcher."), []byte(""), -1)
	b = bytes.Replace(b, []byte("CORE."), []byte(""), -1)
	b = bytes.Replace(b, []byte("SEARCHER."), []byte(""), -1)
	b = bytes.Replace(b, []byte("INDEX."), []byte(""), -1)
	if err := json.Unmarshal(b, &coreMetrics); err != nil {
		errors = append(errors, fmt.Errorf("Failed to unmarshal mbeans core metrics JSON into struct: %v", err))
		return errors
	}

	for name, metrics := range coreMetrics {
		if strings.Contains(name, "@") {
			continue
		}

		e.gaugeCore["deleted_docs"].WithLabelValues(coreName, name, metrics.Class).Set(float64(metrics.Stats.DeletedDocs))
		e.gaugeCore["max_docs"].WithLabelValues(coreName, name, metrics.Class).Set(float64(metrics.Stats.MaxDoc))
		e.gaugeCore["num_docs"].WithLabelValues(coreName, name, metrics.Class).Set(float64(metrics.Stats.NumDocs))
	}

	cacheData := findMBeansData(mBeansData.SolrMbeans, "CACHE")
	b = bytes.Replace(cacheData, []byte(":\"NaN\""), []byte(":0.0"), -1)
	b = bytes.Replace(b, []byte("CACHE.searcher.perSegFilter."), []byte(""), -1)
	b = bytes.Replace(b, []byte("CACHE.searcher.queryResultCache."), []byte(""), -1)
	b = bytes.Replace(b, []byte("CACHE.searcher.fieldValueCache."), []byte(""), -1)
	b = bytes.Replace(b, []byte("CACHE.searcher.filterCache."), []byte(""), -1)
	b = bytes.Replace(b, []byte("CACHE.searcher.documentCache."), []byte(""), -1)
	mbeanerrs := handleCacheMbeans(b, e, coreName)
	for _, e := range mbeanerrs {
		errors = append(errors, e)
	}
	return errors
}

func handleCacheMbeans(data []byte, e *Exporter, coreName string) []error {
	var cacheMetrics map[string]Cache
	var errors = []error{}
	if err := json.Unmarshal(data, &cacheMetrics); err != nil {
		errors = append(errors, fmt.Errorf("Failed to unmarshal mbeans cache metrics JSON into struct (core : %s): %v, json : %s", coreName, err, data))
	} else {
		for name, metrics := range cacheMetrics {
			if metrics.Class == "org.apache.solr.search.SolrFieldCacheMBean" || metrics.Class == "org.apache.solr.search.SolrFieldCacheBean" {
				continue
			}
			hitratio, err := strconv.ParseFloat(string(metrics.Stats.Hitratio), 64)
			if err != nil {
				errors = append(errors, fmt.Errorf("Fail to convert Hitratio in float: %v", err))
			}
			cumulativeHitratio, err := strconv.ParseFloat(string(metrics.Stats.CumulativeHitratio), 64)
			if err != nil {
				errors = append(errors, fmt.Errorf("Fail to convert Cumulative Hitratio in float: %v", err))
			}
			e.gaugeCache["cumulative_evictions"].WithLabelValues(coreName, name, metrics.Class).Set(float64(metrics.Stats.CumulativeEvictions))
			e.gaugeCache["cumulative_hitratio"].WithLabelValues(coreName, name, metrics.Class).Set(cumulativeHitratio)
			e.gaugeCache["cumulative_hits"].WithLabelValues(coreName, name, metrics.Class).Set(float64(metrics.Stats.CumulativeHits))
			e.gaugeCache["cumulative_inserts"].WithLabelValues(coreName, name, metrics.Class).Set(float64(metrics.Stats.CumulativeInserts))
			e.gaugeCache["cumulative_lookups"].WithLabelValues(coreName, name, metrics.Class).Set(float64(metrics.Stats.CumulativeLookups))
			e.gaugeCache["evictions"].WithLabelValues(coreName, name, metrics.Class).Set(float64(metrics.Stats.Evictions))
			e.gaugeCache["hitratio"].WithLabelValues(coreName, name, metrics.Class).Set(hitratio)
			e.gaugeCache["hits"].WithLabelValues(coreName, name, metrics.Class).Set(float64(metrics.Stats.Hits))
			e.gaugeCache["inserts"].WithLabelValues(coreName, name, metrics.Class).Set(float64(metrics.Stats.Inserts))
			e.gaugeCache["lookups"].WithLabelValues(coreName, name, metrics.Class).Set(float64(metrics.Stats.Lookups))
			e.gaugeCache["size"].WithLabelValues(coreName, name, metrics.Class).Set(float64(metrics.Stats.Size))
			e.gaugeCache["warmup_time"].WithLabelValues(coreName, name, metrics.Class).Set(float64(metrics.Stats.WarmupTime))
		}
	}
	return errors
}
