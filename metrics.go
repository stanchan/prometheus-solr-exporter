package main

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/buger/jsonparser"
)

func getCoreAttributes(regExp, coreString string) (coreAttributes map[string]string) {

	var compRegExp = regexp.MustCompile(regExp)
	match := compRegExp.FindStringSubmatch(coreString)

	coreAttributes = make(map[string]string)
	for i, name := range compRegExp.SubexpNames() {
		if i > 0 && i <= len(match) {
			coreAttributes[name] = match[i]
		}
	}
	return coreAttributes
}

func processQueryMetrics(e *Exporter, coreName string, data []byte) []error {
	errors := []error{}
	re := `^(?P<Collection>\w+)_(?P<Shard>shard\d+?)_(?P<Replica>replica_n\d+)$`

	coreMap := getCoreAttributes(re, coreName)
	if coreMap == nil {
		errors = append(errors, fmt.Errorf("Failed to process core re: %v, core: %s", coreMap, coreName))
		return errors
	}
	metricsNode := fmt.Sprintf("solr.core.%s.%s.%s", coreMap["Collection"], coreMap["Shard"], coreMap["Replica"])

	b := bytes.Replace(data, []byte(":\"NaN\""), []byte(":0.0"), -1)
	b = bytes.Replace(b, []byte("ms\","), []byte("\","), -1)

	paths := [][]string{
		[]string{"metrics", metricsNode, "QUERY./select.requestTimes", "15minRate"},
		[]string{"metrics", metricsNode, "QUERY./select.requestTimes", "5minRate"},
		[]string{"metrics", metricsNode, "QUERY./select.requestTimes", "p75_ms"},
		[]string{"metrics", metricsNode, "QUERY./select.requestTimes", "p95_ms"},
		[]string{"metrics", metricsNode, "QUERY./select.requestTimes", "p99_ms"},
		[]string{"metrics", metricsNode, "QUERY./select.requestTimes", "p999_ms"},
		[]string{"metrics", metricsNode, "QUERY./select.requestTimes", "meanRate"},
		[]string{"metrics", metricsNode, "QUERY./select.requestTimes", "mean_ms"},
		[]string{"metrics", metricsNode, "QUERY./select.requestTimes", "median_ms"},
		[]string{"metrics", metricsNode, "QUERY./select.requests"},
		[]string{"metrics", metricsNode, "QUERY./select.errors", "count"},
		[]string{"metrics", metricsNode, "QUERY./select.clientErrors", "count"},
		[]string{"metrics", metricsNode, "QUERY./select.serverErrors", "count"},
		[]string{"metrics", metricsNode, "QUERY./select.handlerStart"},
		[]string{"metrics", metricsNode, "QUERY./select.timeouts", "count"},
		[]string{"metrics", metricsNode, "QUERY./select.totalTime"},
	}

	queryValues := make(map[string]float64)

	jsonparser.EachKey(b, func(idx int, value []byte, vt jsonparser.ValueType, err error) {
		switch idx {
		case 0:
			v, _ := jsonparser.ParseFloat(value)
			queryValues["15min_rate_reqs_per_second"] = v
		case 1:
			v, _ := jsonparser.ParseFloat(value)
			queryValues["5min_rate_reqs_per_second"] = v
		case 2:
			v, _ := jsonparser.ParseFloat(value)
			queryValues["75th_pc_request_time"] = v
		case 3:
			v, _ := jsonparser.ParseFloat(value)
			queryValues["95th_pc_request_time"] = v
		case 4:
			v, _ := jsonparser.ParseFloat(value)
			queryValues["99th_pc_request_time"] = v
		case 5:
			v, _ := jsonparser.ParseFloat(value)
			queryValues["999th_pc_request_time"] = v
		case 6:
			v, _ := jsonparser.ParseFloat(value)
			queryValues["avg_requests_per_second"] = v
		case 7:
			v, _ := jsonparser.ParseFloat(value)
			queryValues["avg_time_per_request"] = v
		case 8:
			v, _ := jsonparser.ParseFloat(value)
			queryValues["median_time_per_request"] = v
		case 9:
			v, _ := jsonparser.ParseFloat(value)
			queryValues["requests"] = v
		case 10:
			v, _ := jsonparser.ParseFloat(value)
			queryValues["errors"] = v
		case 11:
			v, _ := jsonparser.ParseFloat(value)
			queryValues["client_errors"] = v
		case 12:
			v, _ := jsonparser.ParseFloat(value)
			queryValues["server_errors"] = v
		case 13:
			v, _ := jsonparser.ParseFloat(value)
			queryValues["handler_start"] = v
		case 14:
			v, _ := jsonparser.ParseFloat(value)
			queryValues["timeouts"] = v
		case 15:
			v, _ := jsonparser.ParseFloat(value)
			queryValues["total_time"] = v
		}
	}, paths...)

	//fmt.Printf("Processing Query Metrics!: %#v\n", queryValues)

	e.gaugeQuery["15min_rate_reqs_per_second"].WithLabelValues(coreName, "/select.requestTimes", "QUERY").Set(queryValues["15min_rate_reqs_per_second"])
	e.gaugeQuery["5min_rate_reqs_per_second"].WithLabelValues(coreName, "/select.requestTimes", "QUERY").Set(queryValues["5min_rate_reqs_per_second"])
	e.gaugeQuery["75th_pc_request_time"].WithLabelValues(coreName, "/select.requestTimes", "QUERY").Set(queryValues["75th_pc_request_time"])
	e.gaugeQuery["95th_pc_request_time"].WithLabelValues(coreName, "/select.requestTimes", "QUERY").Set(queryValues["95th_pc_request_time"])
	e.gaugeQuery["99th_pc_request_time"].WithLabelValues(coreName, "/select.requestTimes", "QUERY").Set(queryValues["99th_pc_request_time"])
	e.gaugeQuery["999th_pc_request_time"].WithLabelValues(coreName, "/select.requestTimes", "QUERY").Set(queryValues["999th_pc_request_time"])
	e.gaugeQuery["avg_requests_per_second"].WithLabelValues(coreName, "/select.requestTimes", "QUERY").Set(queryValues["avg_requests_per_second"])
	e.gaugeQuery["avg_time_per_request"].WithLabelValues(coreName, "/select.requestTimes", "QUERY").Set(queryValues["avg_time_per_request"])
	e.gaugeQuery["median_time_per_request"].WithLabelValues(coreName, "/select.requestTimes", "QUERY").Set(queryValues["median_time_per_request"])
	e.gaugeQuery["requests"].WithLabelValues(coreName, "/select.requests", "QUERY").Set(queryValues["requests"])
	e.gaugeQuery["errors"].WithLabelValues(coreName, "/select.errors", "QUERY").Set(queryValues["errors"])
	e.gaugeQuery["client_errors"].WithLabelValues(coreName, "/select.clientErrors", "QUERY").Set(queryValues["client_errors"])
	e.gaugeQuery["server_errors"].WithLabelValues(coreName, "/select.serverErrors", "QUERY").Set(queryValues["server_errors"])
	e.gaugeQuery["handler_start"].WithLabelValues(coreName, "/select.handlerStart", "QUERY").Set(queryValues["handler_start"])
	e.gaugeQuery["timeouts"].WithLabelValues(coreName, "/select.timeouts", "QUERY").Set(queryValues["timeouts"])
	e.gaugeQuery["total_time"].WithLabelValues(coreName, "/select.totalTime", "QUERY").Set(queryValues["total_time"])

	return errors
}

func processUpdateMetrics(e *Exporter, coreName string, data []byte) []error {
	errors := []error{}
	re := `^(?P<Collection>\w+)_(?P<Shard>shard\d+?)_(?P<Replica>replica_n\d+)$`

	coreMap := getCoreAttributes(re, coreName)
	if coreMap == nil {
		errors = append(errors, fmt.Errorf("Failed to process core re: %v, core: %s", coreMap, coreName))
		return errors
	}
	metricsNode := fmt.Sprintf("solr.core.%s.%s.%s", coreMap["Collection"], coreMap["Shard"], coreMap["Replica"])

	b := bytes.Replace(data, []byte(":\"NaN\""), []byte(":0.0"), -1)
	b = bytes.Replace(b, []byte("ms\","), []byte("\","), -1)

	paths := [][]string{
		[]string{"metrics", metricsNode, "UPDATE./update.requestTimes", "15minRate"},
		[]string{"metrics", metricsNode, "UPDATE./update.requestTimes", "5minRate"},
		[]string{"metrics", metricsNode, "UPDATE./update.requestTimes", "p75_ms"},
		[]string{"metrics", metricsNode, "UPDATE./update.requestTimes", "p95_ms"},
		[]string{"metrics", metricsNode, "UPDATE./update.requestTimes", "p99_ms"},
		[]string{"metrics", metricsNode, "UPDATE./update.requestTimes", "p999_ms"},
		[]string{"metrics", metricsNode, "UPDATE./update.requestTimes", "meanRate"},
		[]string{"metrics", metricsNode, "UPDATE./update.requestTimes", "mean_ms"},
		[]string{"metrics", metricsNode, "UPDATE./update.requestTimes", "median_ms"},
		[]string{"metrics", metricsNode, "UPDATE./update.requests"},
		[]string{"metrics", metricsNode, "UPDATE.updateHandler.adds"},
		[]string{"metrics", metricsNode, "UPDATE.updateHandler.autoCommitMaxTime"},
		[]string{"metrics", metricsNode, "UPDATE.updateHandler.autoCommits"},
		[]string{"metrics", metricsNode, "UPDATE.updateHandler.commits", "count"},
		[]string{"metrics", metricsNode, "UPDATE.updateHandler.cumulativeAdds", "count"},
		[]string{"metrics", metricsNode, "UPDATE.updateHandler.cumulativeDeletesById", "count"},
		[]string{"metrics", metricsNode, "UPDATE.updateHandler.cumulativeDeletesByQuery", "count"},
		[]string{"metrics", metricsNode, "UPDATE.updateHandler.cumulativeErrors", "count"},
		[]string{"metrics", metricsNode, "UPDATE.updateHandler.deletesById"},
		[]string{"metrics", metricsNode, "UPDATE.updateHandler.deletesByQuery"},
		[]string{"metrics", metricsNode, "UPDATE.updateHandler.docsPending"},
		[]string{"metrics", metricsNode, "UPDATE.updateHandler.errors"},
		[]string{"metrics", metricsNode, "UPDATE.updateHandler.expungeDeletes", "count"},
		[]string{"metrics", metricsNode, "UPDATE.updateHandler.merges", "count"},
		[]string{"metrics", metricsNode, "UPDATE.updateHandler.optimizes", "count"},
		[]string{"metrics", metricsNode, "UPDATE.updateHandler.rollbacks", "count"},
		[]string{"metrics", metricsNode, "UPDATE.updateHandler.softAutoCommits"},
	}

	updateValues := make(map[string]float64)

	jsonparser.EachKey(b, func(idx int, value []byte, vt jsonparser.ValueType, err error) {
		switch idx {
		case 0:
			v, _ := jsonparser.ParseFloat(value)
			updateValues["15min_rate_updates_per_second"] = v
		case 1:
			v, _ := jsonparser.ParseFloat(value)
			updateValues["5min_rate_updates_per_second"] = v
		case 2:
			v, _ := jsonparser.ParseFloat(value)
			updateValues["75th_pc_update_time"] = v
		case 3:
			v, _ := jsonparser.ParseFloat(value)
			updateValues["95th_pc_update_time"] = v
		case 4:
			v, _ := jsonparser.ParseFloat(value)
			updateValues["99th_pc_update_time"] = v
		case 5:
			v, _ := jsonparser.ParseFloat(value)
			updateValues["999th_pc_update_time"] = v
		case 6:
			v, _ := jsonparser.ParseFloat(value)
			updateValues["avg_updates_per_second"] = v
		case 7:
			v, _ := jsonparser.ParseFloat(value)
			updateValues["avg_time_per_update"] = v
		case 8:
			v, _ := jsonparser.ParseFloat(value)
			updateValues["median_time_per_update"] = v
		case 9:
			v, _ := jsonparser.ParseFloat(value)
			updateValues["requests"] = v
		case 10:
			v, _ := jsonparser.ParseFloat(value)
			updateValues["adds"] = v
		case 11:
			v, _ := jsonparser.ParseFloat(value)
			updateValues["autocommit_max_time"] = v
		case 12:
			v, _ := jsonparser.ParseFloat(value)
			updateValues["autocommits"] = v
		case 13:
			v, _ := jsonparser.ParseFloat(value)
			updateValues["commits"] = v
		case 14:
			v, _ := jsonparser.ParseFloat(value)
			updateValues["cumulative_adds"] = v
		case 15:
			v, _ := jsonparser.ParseFloat(value)
			updateValues["cumulative_deletes_by_id"] = v
		case 16:
			v, _ := jsonparser.ParseFloat(value)
			updateValues["cumulative_deletes_by_query"] = v
		case 17:
			v, _ := jsonparser.ParseFloat(value)
			updateValues["cumulative_errors"] = v
		case 18:
			v, _ := jsonparser.ParseFloat(value)
			updateValues["deletes_by_id"] = v
		case 19:
			v, _ := jsonparser.ParseFloat(value)
			updateValues["deletes_by_query"] = v
		case 20:
			v, _ := jsonparser.ParseFloat(value)
			updateValues["docs_pending"] = v
		case 21:
			v, _ := jsonparser.ParseFloat(value)
			updateValues["errors"] = v
		case 22:
			v, _ := jsonparser.ParseFloat(value)
			updateValues["expunge_deletes"] = v
		case 23:
			v, _ := jsonparser.ParseFloat(value)
			updateValues["merges"] = v
		case 24:
			v, _ := jsonparser.ParseFloat(value)
			updateValues["optimizes"] = v
		case 25:
			v, _ := jsonparser.ParseFloat(value)
			updateValues["rollbacks"] = v
		case 26:
			v, _ := jsonparser.ParseFloat(value)
			updateValues["soft_autocommits"] = v
		}
	}, paths...)

	//fmt.Printf("Processing Update Metrics!: %#v\n", updateValues)

	e.gaugeUpdate["15min_rate_updates_per_second"].WithLabelValues(coreName, "/update.requestTimes", "UPDATE").Set(updateValues["15min_rate_updates_per_second"])
	e.gaugeUpdate["5min_rate_updates_per_second"].WithLabelValues(coreName, "/update.requestTimes", "UPDATE").Set(updateValues["5min_rate_updates_per_second"])
	e.gaugeUpdate["75th_pc_update_time"].WithLabelValues(coreName, "/update.requestTimes", "UPDATE").Set(updateValues["75th_pc_update_time"])
	e.gaugeUpdate["95th_pc_update_time"].WithLabelValues(coreName, "/update.requestTimes", "UPDATE").Set(updateValues["95th_pc_update_time"])
	e.gaugeUpdate["99th_pc_update_time"].WithLabelValues(coreName, "/update.requestTimes", "UPDATE").Set(updateValues["99th_pc_update_time"])
	e.gaugeUpdate["999th_pc_update_time"].WithLabelValues(coreName, "/update.requestTimes", "UPDATE").Set(updateValues["999th_pc_update_time"])
	e.gaugeUpdate["avg_updates_per_second"].WithLabelValues(coreName, "/update.requestTimes", "UPDATE").Set(updateValues["avg_updates_per_second"])
	e.gaugeUpdate["avg_time_per_update"].WithLabelValues(coreName, "/update.requestTimes", "UPDATE").Set(updateValues["avg_time_per_update"])
	e.gaugeUpdate["median_time_per_update"].WithLabelValues(coreName, "/update.requestTimes", "UPDATE").Set(updateValues["median_time_per_update"])
	e.gaugeUpdate["requests"].WithLabelValues(coreName, "/update.requests", "UPDATE").Set(updateValues["requests"])
	e.gaugeUpdate["adds"].WithLabelValues(coreName, "updateHandler", "UPDATE").Set(updateValues["adds"])
	e.gaugeUpdate["autocommit_max_time"].WithLabelValues(coreName, "updateHandler", "UPDATE").Set(updateValues["autocommit_max_time"])
	e.gaugeUpdate["autocommits"].WithLabelValues(coreName, "updateHandler", "UPDATE").Set(updateValues["autocommits"])
	e.gaugeUpdate["commits"].WithLabelValues(coreName, "updateHandler", "UPDATE").Set(updateValues["commits"])
	e.gaugeUpdate["cumulative_adds"].WithLabelValues(coreName, "updateHandler", "UPDATE").Set(updateValues["cumulative_adds"])
	e.gaugeUpdate["cumulative_deletes_by_id"].WithLabelValues(coreName, "updateHandler", "UPDATE").Set(updateValues["cumulative_deletes_by_id"])
	e.gaugeUpdate["cumulative_deletes_by_query"].WithLabelValues(coreName, "updateHandler", "UPDATE").Set(updateValues["cumulative_deletes_by_query"])
	e.gaugeUpdate["cumulative_errors"].WithLabelValues(coreName, "updateHandler", "UPDATE").Set(updateValues["cumulative_errors"])
	e.gaugeUpdate["deletes_by_id"].WithLabelValues(coreName, "updateHandler", "UPDATE").Set(updateValues["deletes_by_id"])
	e.gaugeUpdate["deletes_by_query"].WithLabelValues(coreName, "updateHandler", "UPDATE").Set(updateValues["deletes_by_query"])
	e.gaugeUpdate["docs_pending"].WithLabelValues(coreName, "updateHandler", "UPDATE").Set(updateValues["docs_pending"])
	e.gaugeUpdate["errors"].WithLabelValues(coreName, "updateHandler", "UPDATE").Set(updateValues["errors"])
	e.gaugeUpdate["expunge_deletes"].WithLabelValues(coreName, "updateHandler", "UPDATE").Set(updateValues["expunge_deletes"])
	e.gaugeUpdate["merges"].WithLabelValues(coreName, "updateHandler", "UPDATE").Set(updateValues["merges"])
	e.gaugeUpdate["optimizes"].WithLabelValues(coreName, "updateHandler", "UPDATE").Set(updateValues["optimizes"])
	e.gaugeUpdate["rollbacks"].WithLabelValues(coreName, "updateHandler", "UPDATE").Set(updateValues["rollbacks"])
	e.gaugeUpdate["soft_autocommits"].WithLabelValues(coreName, "updateHandler", "UPDATE").Set(updateValues["soft_autocommits"])

	return errors
}
