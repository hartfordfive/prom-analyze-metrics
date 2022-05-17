# Prometheus metrics verifier


### Description
This project is intended as a solution to gain the ability to perform both linting checks and cardinality verifications on metrics endpoints.   Currently, the `promtool` command *should* allow for this although due to some of the logic, any linting warnings will prevent the cardinality analysis from running. (See [issue #10644](https://github.com/prometheus/prometheus/issues/10644))

Please note this was developped relatively quickly as a test and proof of concept, which means the first few releases may contain some bugs and lack functionality.   Any PRs and/or consructive feedback is much appreciated.


### Building

```
make -B
```


### Running

```
prom-metrics-verifier -port=8080 -cachedir=/tmp/prom-analyzer
```
Run with `-h` flag for list of all flags.

