#!/usr/bin/env python

import os
import sys
import time
import re
import string

import requests

graphite_host = "127.0.0.1"
graphite_port = 8000 

def latest_metric_value(metric):
    recent_query = "http://%s:%d/render" % (graphite_host, graphite_port)
    qr = requests.get(recent_query, params={"target": metric, "format": "json", "from": "-3min"})
    if qr.status_code != 200:
        sys.stderr.write("query code error: %d\n" % (qr.status_code))
        return None
    qj = qr.json()
    if len(qj) != 1: return None
    dp = [e[0] for e in qj[0]["datapoints"] if e[0] is not None]
    if len(dp) < 1: return None
    return dp[-1]

def latest_by_regex(regex):
    kv = kv_by_regex(regex)
    kvstr = ["%s ==> %r"%e for e in kv]
    sys.stderr.write("\n\n%s\n" % ("\n".join(kvstr)))

def kv_by_regex(regex):
    available_query = "http://%s:%d/metrics/index.json" % (graphite_host, graphite_port)
    avqr = requests.get(available_query, params={"from": "-1min"})
    metrics = avqr.json()
    matching = [e for e in metrics if (re.match(regex, e) is not None)]
    kv = [(e, latest_metric_value(e)) for e in matching]
    #sys.stderr.write("\n\n%r\n" % ("\n".join(kv)))
    return [e for e in kv if (e[1] is not None)]

def oshinko_query_vars():
    host = os.environ["OSHINKO_REST_SERVICE_HOST"]
    port = os.environ["OSHINKO_REST_SERVICE_PORT"]
    clust = os.environ["OSHINKO_SPARK_CLUSTER"]
    return (host, port, clust)

def oshinko_scale_request(num_workers):
    if num_workers == 0:
        sys.stderr.write("Skipping scale to 0 ...\n")
        return
    host, port, clust = oshinko_query_vars()
    endpoint = "http://%s:%s/clusters/%s" % (host, port, clust)
    #sys.stderr.write("endpoint= %s\n" % (endpoint))
    body = { "name": clust, "config": {"workerCount": int(num_workers)}}
    res = requests.put(endpoint, json = body)
    sys.stderr.write("    scale request status= %d\n" % (res.status_code))

def oshinko_cluster_json():
    host, port, clust = oshinko_query_vars()
    endpoint = "http://%s:%s/clusters/%s" % (host, port, clust)
    res = requests.get(endpoint)
    code = res.status_code
    if (code == 200):
        sys.stderr.write("json= %r\n" % (res.json()))
        return (res.json())['cluster']
    else:
        sys.stderr.write("Bad cluster query, status= %d" % (code))
        return None

while True:
    sys.stderr.write("waiting...\n")
    time.sleep(30)
    #latest_by_regex(".*\.executors\..*")
    #latest_by_regex(".*worker.*")
    #latest_by_regex(".*cores.*")
    #latest_by_regex(".*\.numberTargetExecutors$")
    #latest_by_regex("^master.workers$")
    #available = sum([x[1] for x in kv_by_regex("^master.workers$")])
    cluster = oshinko_cluster_json()
    if cluster is None:
        continue
    available = 0
    if cluster.has_key('config') and cluster['config'].has_key('workerCount'):
        available = cluster['config']['workerCount']
    target = sum([x[1] for x in kv_by_regex(".*\.numberTargetExecutors$")])
    target = max(0, target)
    sys.stderr.write("available= %d   target= %d\n" % (available, target))
    # scale up "immediately", scale down "exponential decay"
    if (target > available):
        oshinko_scale_request(target)
    elif (target < available):
        oshinko_scale_request((target + available) / 2)
