### Example output

```bash
$ om products
  +-----------------------------+-----------------+-----------------+-----------------+
  |            NAME             |    AVAILABLE    |     STAGED      |    DEPLOYED     |
  +-----------------------------+-----------------+-----------------+-----------------+
  | appMetrics                  | 2.0.6-dev.005   | 2.0.6-dev.005   | 2.0.6-dev.005   |
  | cf                          | 2.8.16          | 2.8.16          | 2.8.16          |
  | metric-store                | 1.4.4           | 1.4.4           | 1.4.4           |
  | p-bosh                      |                 | 2.8.2-build.203 | 2.8.2-build.203 |
  | p-event-alerts              | 1.2.9-build.1   | 1.2.9-build.1   | 1.2.9-build.1   |
  | p-healthwatch               | 1.8.3-build.3   | 1.8.3-build.3   | 1.8.2-build.1   |
  | p-healthwatch2              | 2.0.5-build.84  | 2.0.5-build.84  | 2.0.5-build.84  |
  | p-healthwatch2-pas-exporter | 2.0.5-build.84  | 2.0.5-build.84  | 2.0.5-build.84  |
  | p-isolation-segment         | 2.8.2           | 2.8.2           | 2.8.2           |
  | p-rabbitmq                  | 1.19.1-build.22 | 1.19.1-build.22 | 1.19.1-build.22 |
  | p-redis                     | 2.3.1-build.36  | 2.3.1-build.36  | 2.3.1-build.36  |
  | pivotal-mysql               | 2.8.0-build.111 | 2.8.0-build.111 | 2.8.0-build.111 |
  | pivotal-telemetry-om        | 1.0.1-build.2   | 1.0.1-build.2   | 1.0.1-build.2   |
  +-----------------------------+-----------------+-----------------+-----------------+
```

```bash
$ om products --deployed
  +-----------------------------+-----------------+
  |            NAME             |    DEPLOYED     |
  +-----------------------------+-----------------+
  | appMetrics                  | 2.0.6-dev.005   |
  | cf                          | 2.8.16          |
  | metric-store                | 1.4.4           |
  | p-bosh                      | 2.8.2-build.203 |
  | p-event-alerts              | 1.2.9-build.1   |
  | p-healthwatch               | 1.8.2-build.1   |
  | p-healthwatch2              | 2.0.5-build.84  |
  | p-healthwatch2-pas-exporter | 2.0.5-build.84  |
  | p-isolation-segment         | 2.8.2           |
  | p-rabbitmq                  | 1.19.1-build.22 |
  | p-redis                     | 2.3.1-build.36  |
  | pivotal-mysql               | 2.8.0-build.111 |
  | pivotal-telemetry-om        | 1.0.1-build.2   |
  +-----------------------------+-----------------+
```
