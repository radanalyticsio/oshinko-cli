#!/usr/bin/env python

import os
import sys
import time
import re
import string

import requests

graphite_host = "127.0.0.1"
graphite_port = 8000 

poll_interval = 5
poll_window = 30
window_len = poll_window / poll_interval

def latest_metric_value(metric):
    recent_query = "http://%s:%d/render" % (graphite_host, graphite_port)
    qr = requests.get(recent_query, params={"target": metric, "format": "json", "from": "-1min"})
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
    avqr = requests.get(available_query)
    metrics = avqr.json()
    matching = [e for e in metrics if (re.match(regex, e) is not None)]
    if len(matching) < 1:
        sys.stderr.write("No published metrics matching regex: %s\n" % (regex))
    kv = [(e, latest_metric_value(e)) for e in matching]
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
    body = { "name": clust, "config": {"workerCount": int(num_workers)}}
    res = requests.put(endpoint, json = body)
    sys.stderr.write("    scale request status= %d\n" % (res.status_code))

def oshinko_cluster_json():
    host, port, clust = oshinko_query_vars()
    endpoint = "http://%s:%s/clusters/%s" % (host, port, clust)
    res = requests.get(endpoint)
    code = res.status_code
    if (code == 200):
        return (res.json())['cluster']
    else:
        sys.stderr.write("Bad cluster query, status= %d\n" % (code))
        return None

sys.stderr.write("started elastic worker daemon.\n")

target_window = []

while True:
    time.sleep(poll_interval)
    cluster = oshinko_cluster_json()
    if cluster is None:
        sys.stderr.write("No cluster...\n")
        continue
    available = 0
    if cluster.has_key('config') and cluster['config'].has_key('workerCount'):
        available = cluster['config']['workerCount']
    qt = [x[1] for x in kv_by_regex(".*\.numberTargetExecutors$")]
    target = sum(qt)
    target = max(1, target)
    if len(target_window) >= window_len:
        target_window = target_window[1:]
    target_window.append(target)
    sys.stderr.write("recent executor requests: %s\n" % (target_window))
    target = max(target_window)
    # scale up "immediately", scale down "exponential decay"
    if (target > available):
        sys.stderr.write("scaling up from %d to %d\n" % (available, target))
        oshinko_scale_request(target)
    elif (target < available):
        target = (target + available) / 2
        sys.stderr.write("scaling down from %d to %d\n" % (available, target))
        oshinko_scale_request(target)
