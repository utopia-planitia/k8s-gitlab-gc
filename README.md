# k8s-gitlab-gc

## supported annotations for k8s namespace manifest/resource

| (default) keys                                    | value type | value examples |
|---------------------------------------------------|:-----------:|-------------:|
|`"utopia-planitia/k8s-gitlab-gc/disable-automatic-garbage-collection"`| string bool | `"true"` (every other string will be evaluated as false) |
| `"utopia-planitia/k8s-gitlab-gc/ns-ttl-duration"` |  string duration (uses go's ParseDuration function, which means valid time units are 'ns', 'us' (or 'µs'), 'ms', 's', 'm', 'h'.) | `"30m"` or `"2h45m"` |

> Note: some name's of keys can be configured (overwritten) via command line flags, e.g. for the `ttlAnnotation` which has a defaukt key like `utopia-planitia/k8s-gitlab-gc/ns-ttl-duration` but can be overwritten
