# analytics-pipeline-monitor

diagnoses latency-related metrics issues

Owned by eng-internal-products

## Overview

`analytics-pipeline-monitor` (APM) runs as a worker that connects to Redshift on an hourly basis and performs the following functions:

- Surfaces data latency by querying all tables for the latest data timestamp. A recent timestamp means SOME data is fresh; an old timestamp means NO data is fresh.
- Delivers actionable alerts by posting a message in #oncall-ip whenever the data latency exceeds a configurable per-table threshold (ex. 2h).
- Easily improves coverage by declaring latency checks and thresholds in a config file. No application code is required to add additional checks or alerts for future tables.
- Default values for latency and timestamp column can be set per schema. This means that if new tables are added to a schema already checked by APM, then APM will latency check those tables using the default values without requiring any config changes. If the new table has characteristics differing from the schema defaults, then they can be overriden by making a config change (shown below).

## Declaring New Latency Checks
Defining checks in analytics-pipeline-monitor can be accomplished by adding a new entry to `config/latency_config.json` in the following format:

```
  "prod": [
    {
      "schema": "mongo",
      "default_threshold": "24h",
      "default_timestamp_column": "_data_timestamp",
      "omit_tables": ["billing_03_31_snapshot"],
      "checks": [
        {
          "table": "districts",
          "latency": {
            "timestamp_column": "_data_timestamp",
            "threshold": "2h",
          }
        },
      ]
    }, ...
  ]
```

`analytics-pipeline-monitor` then reads from this config to perform latency checks. `schema` + `table` identifies the table, and `latency.timestamp_column` identifies the time a row enters Redshift. `latency.threshold` configures the maximum amount of latency acceptable for the table's data in [Go time format](https://golang.org/pkg/time/#ParseDuration). If the threshold is exceeded, then analytics-pipeline-monitor fires an alert in SignalFx.

For tables that are not explicitly declared in the config, `default_threshold` and `default_timestamp_column` will be used as substitutes for the above values. `omit_tables` allows tables to be whitelisted from latency checks.

## Running locally

```
ark start analytics-pipeline-monitor -e clever-dev -l
```

## Deploying

```
ark start analytics-pipeline-monitor -e production
```
