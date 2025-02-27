# benchmarker
SuperMassive benchmark tool

## Build it
Go 1.24 required..

```bash
go build -o supermassive-benchmark
```

| Flag               | Description                                      | Default Value   |
|--------------------|--------------------------------------------------|-----------------|
| `--cluster`        | Cluster address (e.g., localhost:4000)           | `localhost:4000`|
| `--user`           | Username for authentication                      | `username`      |
| `--pass`           | Password for authentication                      | `password`      |
| `--workers`        | Number of workers (goroutines)                   | `8`             |
| `--operations`     | Total number of operations to perform            | `1000000`       |
| `--keysize`        | Size of the key in bytes                         | `16`            |
| `--valuesize`      | Size of the value in bytes                       | `16`            |
| `--readratio`      | Ratio of read operations                         | `0.8`           |
| `--writeratio`     | Ratio of write operations                        | `0.15`          |
| `--deleteratio`    | Ratio of delete operations                       | `0.05`          |
| `--tls`            | Use TLS for connections                          | `false`         |
| `--tls-cert`       | Path to TLS certificate file                     | `""`            |
| `--tls-key`        | Path to TLS key file                             | `""`            |
| `--tls-ca-cert`    | Path to CA certificate file for verification     | `""`            |
| `--tls-skip-verify`| Skip TLS certificate verification (insecure)     | `false`         |
| `--tls-server-name`| Server name for TLS verification                 | `""`            |