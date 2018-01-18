FROM debian:jessie
RUN apt-get update -y && \
    apt-get install -y ca-certificates curl && \
    curl -L https://github.com/Clever/gearcmd/releases/download/0.10.0/gearcmd-v0.10.0-linux-amd64.tar.gz | tar xz -C /usr/local/bin --strip-components 1

COPY bin/analytics-pipeline-monitor /usr/bin/analytics-pipeline-monitor
COPY config/latency_config.json /usr/bin/config/latency_config.json
COPY kvconfig.yml /usr/bin/kvconfig.yml

CMD ["gearcmd", "--name", "analytics-pipeline-monitor", "--cmd", "/usr/bin/analytics-pipeline-monitor", "--parseargs=false"]
