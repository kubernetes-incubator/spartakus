# Spartakus

This project aims to collect information about Kubernetes clusters.  This
information will help the Kubernetes development team to understand what people
are doing with it, and to make it a better product.

THIS DOES NOT REPORT ANY PERSONAL INFORMATION.  Anything that might be
identifying, including things like IP addresses, container images, and object
names are anonymized.  We take this very seriously.  If you think something we
are collecting might be considered personal information, PLEASE let us know!

This is a 100% voluntary effort.  It is not baked into Kubernetes in any way,
shape, or form.  If you don't want to run it, you don't have to.  If you want
to run your own server and collect data yourself, you can do that, too.  If you
want to report less information, the source is here and we're open to feature
requests to better tailor the reports to what people are comfortable with.

## Overview

This project encompasses two things:

1. collector: an HTTP API which receives reports from volunteers and stores
   them in Google Bigquery
2. volunteer: a tool that periodically generates reports using the Kubernetes
   API and publishes it to the collector

## What is in a report?

Reports include a user-provided cluster identifier (we recommend a random
UUID), the version strings of your kubernetes master, and some information
about each node in the cluster, including OS version, kubelet version, docker
version, and CPU and memory capacity.  Total number of namespaces in a k8's 
cluster. Average number of pods, jobs, services, deployments and average 
lifetime per namespace. 

To repeat from above: THIS DOES NOT REPORT ANY PERSONAL INFORMATION.  Anything
that might be identifying, including things like IP addresses, container
images, and object names are anonymized.  We take this very seriously.  If you
think something we are collecting might be considered personal information,
PLEASE let us know!

An example of a report payload:

```json
{
  "version": "v1.0.0",
  "timestamp": "867530909031976",
  "clusterID": "2f9c93d3-156c-47aa-8802-578ffca9b50e",
  "masterVersion": "v1.3.5",
  "nodes": [
    {
      "id": "c8863d09ecc5be8d9791f72acd275fc2",
      "operatingSystem": "linux",
      "osImage": "Debian GNU/Linux 7 (wheezy)",
      "kernelVersion": "3.16.0-4-amd64",
      "architecture": "amd64",
      "containerRuntimeVersion": "docker://1.11.2",
      "kubeletVersion": "v1.3.2",
      "capacity": [
        {
          "resource": "alpha.kubernetes.io/nvidia-gpu",
          "value": "0"
        },
        {
          "resource": "cpu",
          "value": "4"
        },
        {
          "resource": "memory",
          "value": "15437428Ki"
        },
        {
          "resource": "pods",
          "value": "110"
        }
      ]
    },
    {
      "id": "5b919a15947b0680277acddf68d4b7aa",
      "operatingSystem": "linux",
      "osImage": "Debian GNU/Linux 7 (wheezy)",
      "kernelVersion": "3.16.0-4-amd64",
      "architecture": "amd64",
      "containerRuntimeVersion": "docker://1.11.2",
      "kubeletVersion": "v1.3.2",
      "capacity": [
        {
          "resource": "alpha.kubernetes.io/nvidia-gpu",
          "value": "0"
        },
        {
          "resource": "cpu",
          "value": "4"
        },
        {
          "resource": "memory",
          "value": "15437428Ki"
        },
        {
          "resource": "pods",
          "value": "110"
        }
      ]
    }
  ],
  "namespaces": {
    "count": 8,
    "lifetime": {
      "count": 180,
      "avg": 141.2,
      "min": 1,
      "max": 180,
      "histogram": [
        {
          "bucket": 1,
          "cumulative_count": 45,
          "lower_bound": 0,
          "upper_bound": 5,
          "percentage": 45
        },
        {
          "bucket": 2,
          "cumulative_count": 2,
          "lower_bound": 5,
          "upper_bound": 10,
          "percentage": 15
        },
        {
          "bucket": 3,
          "cumulative_count": 23,
          "lower_bound": 10,
          "upper_bound": "inf",
          "percentage": 19
        }
      ]
    },
    "pods": {
      "count": 45,
      "avg": 141.2,
      "min": 1,
      "max": 180,
      "histogram": [
        {
          "bucket": 1,
          "cumulative_count": 45,
          "lower_bound": 0,
          "upper_bound": 5,
          "percentage": 34
        },
        {
          "bucket": 2,
          "cumulative_count": 2,
          "lower_bound": 0,
          "upper_bound": 10,
          "percentage": 50
        },
        {
          "bucket": 3,
          "cumulative_count": 23,
          "lower_bound": 0,
          "upper_bound": "inf",
          "percentage": 14
        }
      ]
    }
  }
}
```

### Future

We'd like to add more information to reports, but we're starting small.
Anything we add will follow the same strict privacy rules outlined above.  Some
examples of things we would consider adding:
   * How many Namespaces exist
   * A histogram of how many Pods, Services, Deployments, etc. exist
     per-Namespace
   * Average lifetime of Namespaces, Pods, Services, etc.

## What will we do with this information?

This information, in aggregate, will help the Kubernetes development teams
prioritize our efforts.  The better we understand what people are doing, the
better we can focus on the important issues.

## How do I run it?

The simplest way is to use `kubectl run`.

```
kubectl run spartakus \
    --image=gcr.io/google_containers/spartakus-amd64:v1.0.0 \
    -- \
    volunteer --cluster-id=$(uuidgen -r)
```

This will generate a `Deployment` called "spartakus" in your "default"
`Namespace`, which sends a report once every 24 hours.  It will report a random
UUID as the cluster ID.  If you ever need to stop this deployment and re-run
it, the UUID will be different.  If you want to generate a more durable UUID by
hand, that's fine too.

If you want to save the YAML that this produces, you can simply run:

```
kubectl get deployment spartakus -o yaml
```

## Development

Simply run `make` or `make test`.  The build is done through Docker.
