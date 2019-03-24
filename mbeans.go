package main

import (
	"bytes"
	"fmt"

	"github.com/buger/jsonparser"
)

func processMbeans(e *Exporter, coreName string, data []byte) []error {
	errors := []error{}

	b := bytes.Replace(data, []byte(":\"NaN\""), []byte(":0.0"), -1)
	b = bytes.Replace(b, []byte("ms\","), []byte("\","), -1)

	i := 0
	jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if string(value) == "CORE" {
			coreerrs := handleCoreMbeans(b, e, coreName, i+1)
			for _, e := range coreerrs {
				errors = append(errors, e)
			}
		} else if string(value) == "CACHE" {
			coreerrs := handleCacheMbeans(b, e, coreName, i+1)
			for _, e := range coreerrs {
				errors = append(errors, e)
			}
		}
		i++
	}, "solr-mbeans")

	return errors
}

func handleCoreMbeans(data []byte, e *Exporter, coreName string, index int) []error {
	coreIdx := fmt.Sprintf("[%d]", index)
	coreValues := make(map[string]float64)
	errors := []error{}

	paths := [][]string{
		[]string{"solr-mbeans", coreIdx, "searcher", "stats", "SEARCHER.searcher.deletedDocs"},
		[]string{"solr-mbeans", coreIdx, "searcher", "stats", "SEARCHER.searcher.maxDoc"},
		[]string{"solr-mbeans", coreIdx, "searcher", "stats", "SEARCHER.searcher.numDocs"},
	}

	jsonparser.EachKey(data, func(idx int, value []byte, vt jsonparser.ValueType, err error) {
		switch idx {
		case 0:
			v, _ := jsonparser.ParseFloat(value)
			coreValues["deleted_docs"] = v
		case 1:
			v, _ := jsonparser.ParseFloat(value)
			coreValues["max_docs"] = v
		case 2:
			v, _ := jsonparser.ParseFloat(value)
			coreValues["num_docs"] = v
		}
	}, paths...)

	e.gaugeCore["deleted_docs"].WithLabelValues(coreName, "searcher", "CORE").Set(coreValues["deleted_docs"])
	e.gaugeCore["max_docs"].WithLabelValues(coreName, "searcher", "CORE").Set(coreValues["max_docs"])
	e.gaugeCore["num_docs"].WithLabelValues(coreName, "searcher", "CORE").Set(coreValues["num_docs"])

	return errors
}

func handleCacheMbeans(data []byte, e *Exporter, coreName string, index int) []error {
	cacheIdx := fmt.Sprintf("[%d]", index)
	cacheValues := map[string]map[string]float64{
		"perSegFilter":     map[string]float64{},
		"queryResultCache": map[string]float64{},
		"fieldValueCache":  map[string]float64{},
		"filterCache":      map[string]float64{},
		"documentCache":    map[string]float64{},
	}
	errors := []error{}

	paths := [][]string{
		[]string{"solr-mbeans", cacheIdx, "perSegFilter", "stats", "CACHE.searcher.perSegFilter.lookups"},
		[]string{"solr-mbeans", cacheIdx, "perSegFilter", "stats", "CACHE.searcher.perSegFilter.hits"},
		[]string{"solr-mbeans", cacheIdx, "perSegFilter", "stats", "CACHE.searcher.perSegFilter.hitratio"},
		[]string{"solr-mbeans", cacheIdx, "perSegFilter", "stats", "CACHE.searcher.perSegFilter.warmupTime"},
		[]string{"solr-mbeans", cacheIdx, "perSegFilter", "stats", "CACHE.searcher.perSegFilter.cumulative_lookups"},
		[]string{"solr-mbeans", cacheIdx, "perSegFilter", "stats", "CACHE.searcher.perSegFilter.cumulative_inserts"},
		[]string{"solr-mbeans", cacheIdx, "perSegFilter", "stats", "CACHE.searcher.perSegFilter.cumulative_hitratio"},
		[]string{"solr-mbeans", cacheIdx, "perSegFilter", "stats", "CACHE.searcher.perSegFilter.size"},
		[]string{"solr-mbeans", cacheIdx, "perSegFilter", "stats", "CACHE.searcher.perSegFilter.evictions"},
		[]string{"solr-mbeans", cacheIdx, "perSegFilter", "stats", "CACHE.searcher.perSegFilter.cumulative_hits"},
		[]string{"solr-mbeans", cacheIdx, "perSegFilter", "stats", "CACHE.searcher.perSegFilter.cumulative_evictions"},
		[]string{"solr-mbeans", cacheIdx, "perSegFilter", "stats", "CACHE.searcher.perSegFilter.inserts"},
		[]string{"solr-mbeans", cacheIdx, "queryResultCache", "stats", "CACHE.searcher.queryResultCache.lookups"},
		[]string{"solr-mbeans", cacheIdx, "queryResultCache", "stats", "CACHE.searcher.queryResultCache.hits"},
		[]string{"solr-mbeans", cacheIdx, "queryResultCache", "stats", "CACHE.searcher.queryResultCache.hitratio"},
		[]string{"solr-mbeans", cacheIdx, "queryResultCache", "stats", "CACHE.searcher.queryResultCache.warmupTime"},
		[]string{"solr-mbeans", cacheIdx, "queryResultCache", "stats", "CACHE.searcher.queryResultCache.cumulative_lookups"},
		[]string{"solr-mbeans", cacheIdx, "queryResultCache", "stats", "CACHE.searcher.queryResultCache.cumulative_inserts"},
		[]string{"solr-mbeans", cacheIdx, "queryResultCache", "stats", "CACHE.searcher.queryResultCache.cumulative_hitratio"},
		[]string{"solr-mbeans", cacheIdx, "queryResultCache", "stats", "CACHE.searcher.queryResultCache.size"},
		[]string{"solr-mbeans", cacheIdx, "queryResultCache", "stats", "CACHE.searcher.queryResultCache.evictions"},
		[]string{"solr-mbeans", cacheIdx, "queryResultCache", "stats", "CACHE.searcher.queryResultCache.cumulative_hits"},
		[]string{"solr-mbeans", cacheIdx, "queryResultCache", "stats", "CACHE.searcher.queryResultCache.cumulative_evictions"},
		[]string{"solr-mbeans", cacheIdx, "queryResultCache", "stats", "CACHE.searcher.queryResultCache.inserts"},
		[]string{"solr-mbeans", cacheIdx, "fieldValueCache", "stats", "CACHE.searcher.fieldValueCache.lookups"},
		[]string{"solr-mbeans", cacheIdx, "fieldValueCache", "stats", "CACHE.searcher.fieldValueCache.hits"},
		[]string{"solr-mbeans", cacheIdx, "fieldValueCache", "stats", "CACHE.searcher.fieldValueCache.hitratio"},
		[]string{"solr-mbeans", cacheIdx, "fieldValueCache", "stats", "CACHE.searcher.fieldValueCache.warmupTime"},
		[]string{"solr-mbeans", cacheIdx, "fieldValueCache", "stats", "CACHE.searcher.fieldValueCache.cumulative_lookups"},
		[]string{"solr-mbeans", cacheIdx, "fieldValueCache", "stats", "CACHE.searcher.fieldValueCache.cumulative_inserts"},
		[]string{"solr-mbeans", cacheIdx, "fieldValueCache", "stats", "CACHE.searcher.fieldValueCache.cumulative_hitratio"},
		[]string{"solr-mbeans", cacheIdx, "fieldValueCache", "stats", "CACHE.searcher.fieldValueCache.size"},
		[]string{"solr-mbeans", cacheIdx, "fieldValueCache", "stats", "CACHE.searcher.fieldValueCache.evictions"},
		[]string{"solr-mbeans", cacheIdx, "fieldValueCache", "stats", "CACHE.searcher.fieldValueCache.cumulative_hits"},
		[]string{"solr-mbeans", cacheIdx, "fieldValueCache", "stats", "CACHE.searcher.fieldValueCache.cumulative_evictions"},
		[]string{"solr-mbeans", cacheIdx, "fieldValueCache", "stats", "CACHE.searcher.fieldValueCache.inserts"},
		[]string{"solr-mbeans", cacheIdx, "filterCache", "stats", "CACHE.searcher.filterCache.lookups"},
		[]string{"solr-mbeans", cacheIdx, "filterCache", "stats", "CACHE.searcher.filterCache.hits"},
		[]string{"solr-mbeans", cacheIdx, "filterCache", "stats", "CACHE.searcher.filterCache.hitratio"},
		[]string{"solr-mbeans", cacheIdx, "filterCache", "stats", "CACHE.searcher.filterCache.warmupTime"},
		[]string{"solr-mbeans", cacheIdx, "filterCache", "stats", "CACHE.searcher.filterCache.cumulative_lookups"},
		[]string{"solr-mbeans", cacheIdx, "filterCache", "stats", "CACHE.searcher.filterCache.cumulative_inserts"},
		[]string{"solr-mbeans", cacheIdx, "filterCache", "stats", "CACHE.searcher.filterCache.cumulative_hitratio"},
		[]string{"solr-mbeans", cacheIdx, "filterCache", "stats", "CACHE.searcher.filterCache.size"},
		[]string{"solr-mbeans", cacheIdx, "filterCache", "stats", "CACHE.searcher.filterCache.evictions"},
		[]string{"solr-mbeans", cacheIdx, "filterCache", "stats", "CACHE.searcher.filterCache.cumulative_hits"},
		[]string{"solr-mbeans", cacheIdx, "filterCache", "stats", "CACHE.searcher.filterCache.cumulative_evictions"},
		[]string{"solr-mbeans", cacheIdx, "filterCache", "stats", "CACHE.searcher.filterCache.inserts"},
		[]string{"solr-mbeans", cacheIdx, "documentCache", "stats", "CACHE.searcher.documentCache.lookups"},
		[]string{"solr-mbeans", cacheIdx, "documentCache", "stats", "CACHE.searcher.documentCache.hits"},
		[]string{"solr-mbeans", cacheIdx, "documentCache", "stats", "CACHE.searcher.documentCache.hitratio"},
		[]string{"solr-mbeans", cacheIdx, "documentCache", "stats", "CACHE.searcher.documentCache.warmupTime"},
		[]string{"solr-mbeans", cacheIdx, "documentCache", "stats", "CACHE.searcher.documentCache.cumulative_lookups"},
		[]string{"solr-mbeans", cacheIdx, "documentCache", "stats", "CACHE.searcher.documentCache.cumulative_inserts"},
		[]string{"solr-mbeans", cacheIdx, "documentCache", "stats", "CACHE.searcher.documentCache.cumulative_hitratio"},
		[]string{"solr-mbeans", cacheIdx, "documentCache", "stats", "CACHE.searcher.documentCache.size"},
		[]string{"solr-mbeans", cacheIdx, "documentCache", "stats", "CACHE.searcher.documentCache.evictions"},
		[]string{"solr-mbeans", cacheIdx, "documentCache", "stats", "CACHE.searcher.documentCache.cumulative_hits"},
		[]string{"solr-mbeans", cacheIdx, "documentCache", "stats", "CACHE.searcher.documentCache.cumulative_evictions"},
		[]string{"solr-mbeans", cacheIdx, "documentCache", "stats", "CACHE.searcher.documentCache.inserts"},
	}

	jsonparser.EachKey(data, func(idx int, value []byte, vt jsonparser.ValueType, err error) {
		switch idx {
		case 0:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["perSegFilter"]["lookups"] = v
		case 1:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["perSegFilter"]["hits"] = v
		case 2:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["perSegFilter"]["hitratio"] = v
		case 3:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["perSegFilter"]["warmup_time"] = v
		case 4:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["perSegFilter"]["cumulative_lookups"] = v
		case 5:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["perSegFilter"]["cumulative_inserts"] = v
		case 6:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["perSegFilter"]["cumulative_hitratio"] = v
		case 7:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["perSegFilter"]["size"] = v
		case 8:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["perSegFilter"]["evictions"] = v
		case 9:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["perSegFilter"]["cumulative_hits"] = v
		case 10:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["perSegFilter"]["cumulative_evictions"] = v
		case 11:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["perSegFilter"]["inserts"] = v
		case 12:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["queryResultCache"]["lookups"] = v
		case 13:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["queryResultCache"]["hits"] = v
		case 14:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["queryResultCache"]["hitratio"] = v
		case 15:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["queryResultCache"]["warmup_time"] = v
		case 16:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["queryResultCache"]["cumulative_lookups"] = v
		case 17:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["queryResultCache"]["cumulative_inserts"] = v
		case 18:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["queryResultCache"]["cumulative_hitratio"] = v
		case 19:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["queryResultCache"]["size"] = v
		case 20:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["queryResultCache"]["evictions"] = v
		case 21:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["queryResultCache"]["cumulative_hits"] = v
		case 22:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["queryResultCache"]["cumulative_evictions"] = v
		case 23:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["queryResultCache"]["inserts"] = v
		case 24:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["fieldValueCache"]["lookups"] = v
		case 25:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["fieldValueCache"]["hits"] = v
		case 26:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["fieldValueCache"]["hitratio"] = v
		case 27:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["fieldValueCache"]["warmup_time"] = v
		case 28:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["fieldValueCache"]["cumulative_lookups"] = v
		case 29:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["fieldValueCache"]["cumulative_inserts"] = v
		case 30:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["fieldValueCache"]["cumulative_hitratio"] = v
		case 31:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["fieldValueCache"]["size"] = v
		case 32:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["fieldValueCache"]["evictions"] = v
		case 33:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["fieldValueCache"]["cumulative_hits"] = v
		case 34:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["fieldValueCache"]["cumulative_evictions"] = v
		case 35:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["fieldValueCache"]["inserts"] = v
		case 36:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["filterCache"]["lookups"] = v
		case 37:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["filterCache"]["hits"] = v
		case 38:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["filterCache"]["hitratio"] = v
		case 39:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["filterCache"]["warmup_time"] = v
		case 40:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["filterCache"]["cumulative_lookups"] = v
		case 41:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["filterCache"]["cumulative_inserts"] = v
		case 42:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["filterCache"]["cumulative_hitratio"] = v
		case 43:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["filterCache"]["size"] = v
		case 44:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["filterCache"]["evictions"] = v
		case 45:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["filterCache"]["cumulative_hits"] = v
		case 46:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["filterCache"]["cumulative_evictions"] = v
		case 47:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["filterCache"]["inserts"] = v
		case 48:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["documentCache"]["lookups"] = v
		case 49:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["documentCache"]["hits"] = v
		case 50:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["documentCache"]["hitratio"] = v
		case 51:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["documentCache"]["warmup_time"] = v
		case 52:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["documentCache"]["cumulative_lookups"] = v
		case 53:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["documentCache"]["cumulative_inserts"] = v
		case 54:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["documentCache"]["cumulative_hitratio"] = v
		case 55:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["documentCache"]["size"] = v
		case 56:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["documentCache"]["evictions"] = v
		case 57:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["documentCache"]["cumulative_hits"] = v
		case 58:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["documentCache"]["cumulative_evictions"] = v
		case 59:
			v, _ := jsonparser.ParseFloat(value)
			cacheValues["documentCache"]["inserts"] = v
		}
	}, paths...)

	//fmt.Printf("Processing Cache Metrics!: %#v\n", cacheValues)

	for element := range cacheValues {
		name := fmt.Sprintf("%s.%s", "searcher", element)
		e.gaugeCache["lookups"].WithLabelValues(coreName, name, "CACHE").Set(cacheValues[element]["lookups"])
		e.gaugeCache["hits"].WithLabelValues(coreName, name, "CACHE").Set(cacheValues[element]["hits"])
		e.gaugeCache["hitratio"].WithLabelValues(coreName, name, "CACHE").Set(cacheValues[element]["hitratio"])
		e.gaugeCache["warmup_time"].WithLabelValues(coreName, name, "CACHE").Set(cacheValues[element]["warmup_time"])
		e.gaugeCache["cumulative_lookups"].WithLabelValues(coreName, name, "CACHE").Set(cacheValues[element]["cumulative_lookups"])
		e.gaugeCache["cumulative_inserts"].WithLabelValues(coreName, name, "CACHE").Set(cacheValues[element]["cumulative_inserts"])
		e.gaugeCache["cumulative_hitratio"].WithLabelValues(coreName, name, "CACHE").Set(cacheValues[element]["cumulative_hitratio"])
		e.gaugeCache["size"].WithLabelValues(coreName, name, "CACHE").Set(cacheValues[element]["size"])
		e.gaugeCache["evictions"].WithLabelValues(coreName, name, "CACHE").Set(cacheValues[element]["evictions"])
		e.gaugeCache["cumulative_hits"].WithLabelValues(coreName, name, "CACHE").Set(cacheValues[element]["cumulative_hits"])
		e.gaugeCache["cumulative_evictions"].WithLabelValues(coreName, name, "CACHE").Set(cacheValues[element]["cumulative_evictions"])
		e.gaugeCache["inserts"].WithLabelValues(coreName, name, "CACHE").Set(cacheValues[element]["inserts"])
	}
	return errors
}
