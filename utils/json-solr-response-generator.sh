#!/bin/bash

solr_tags=$(curl -s https://hub.docker.com/v2/repositories/library/solr/tags/?page_size=100 | jq -r '.results|.[]|.name' | egrep '^(7|8)' | grep -v slim | grep -v alpine | grep -v latest)

for j in ${solr_tags}; do
    docker rm -f solr-$j
    echo "Generate json for solr tag : $j"
    nohup docker run --name solr-$j -p8983:8983 solr:$j &
    until $(curl --output /dev/null --silent --head --fail "http://localhost:8983/solr/#/"); do
        echo "Solr is unavailable - sleeping"
        sleep 1
    done
    echo "Solr $j ready!"
    sleep 2
    docker exec -ti solr-$j bin/solr create_core -c gettingstarted
    sleep 2
    until $(curl --output /dev/null --silent --head --fail "http://localhost:8983/solr/gettingstarted/admin/mbeans?stats=true&wt=json&cat=CORE&cat=QUERYHANDLER&cat=UPDATEHANDLER&cat=CACHE"); do
        echo "Core gettingstarted stats are unavailable - sleeping"
        sleep 1
    done
    mkdir solr-responses/$j || true

    curl --silent --fail "http://localhost:8983/solr/admin/cores?action=STATUS&wt=json" | python -m json.tool > solr-responses/$j/admin-cores.json
    curl --silent --fail "http://localhost:8983/solr/gettingstarted/admin/mbeans?stats=true&wt=json&cat=CORE&cat=QUERYHANDLER&cat=UPDATEHANDLER&cat=CACHE" | python -m json.tool > solr-responses/$j/mbeans.json
    curl --silent --fail "http://localhost:8983/solr/admin/metrics?group=core&prefix=QUERY,UPDATE&wt=json" | python -m json.tool > solr-responses/$j/metrics.json
    docker stop solr-$j
done