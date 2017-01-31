#!/usr/bin/env python
import os
import random
import requests
import string
import sys
import time

def latest_metric_value(metrics_ip, metric):
    recent_query = "http://%s/render" % metrics_ip
    qr = requests.get(recent_query, params={"target": metric, "format": "json", "from": "-1min"})
    if qr.status_code != 200:
        sys.stderr.write("query code error: %d\n" % (qr.status_code))
        return None
    qj = qr.json()
    if len(qj) != 1: return None
    dp = [e[0] for e in qj[0]["datapoints"] if e[0] is not None]
    if len(dp) < 1: return None
    return dp[-1]

def wait_for_workers(cluster, metrics, desired_workers):
    while True:
        workers = latest_metric_value(metrics, "master.aliveWorkers")
        if workers != None and workers >= desired_workers:
            break
        sys.stdout.write("Waiting for workers\n")
        time.sleep(1)
    sys.stdout.write("%d workers alive\n" % desired_workers)

def get_metrics_route(clustername):
    route_suffix = os.environ.get("METRICS_ROUTE_SUFFIX")
    if route_suffix != None:
        return clustername + "-metrics-"+ route_suffix
    return clustername + "-metrics:8000"
    
def create_cluster(osh_ipaddr, clustername, clusterconfig):
    r = requests.post("http://%s/clusters" % osh_ipaddr,
                      json={"name": clustername, "config": {"name": clusterconfig}})
    if r.status_code != 201:
        sys.stderr.write("Cluster creation failed: %d %s\n" % (r.status_code, r.text))
        return False
    json = r.json()
    workers = json['cluster']['config']['workerCount']
    name = json['cluster']['name']
    route = get_metrics_route(name)
    sys.stderr.write("Created cluster %s" % name)
    wait_for_workers(name, route, workers)
    return True

def find_oshinko():
    host = os.environ.get("OSHINKO_REST_SERVICE_HOST")
    port = os.environ.get("OSHINKO_REST_SERVICE_PORT")
    if host == None or port == None:
        sys.stderr.write("Can't determine oshinko host")
        return ""
    return host + ":" + port

def clustername():
    name = os.environ.get("OSHINKO_CLUSTER_NAME")
    if name == None:
        name = "cluster-" + ''.join(random.SystemRandom().choice(string.ascii_lowercase + string.digits) for _ in range(4))
    return name

success = create_cluster(find_oshinko(), clustername(), "clusterconfig")
print(success)
    

