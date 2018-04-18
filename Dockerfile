FROM debian:jessie
RUN apt-get update -y && \
    apt-get install -y ca-certificates

COPY bin/sfncli /usr/bin/sfncli
COPY bin/analytics-monitor /usr/bin/analytics-monitor
COPY config/latency_config.json /usr/bin/config/latency_config.json
COPY kvconfig.yml /usr/bin/kvconfig.yml

CMD ["/usr/bin/sfncli", "--cmd", "/usr/bin/analytics-monitor", "--activityname", "${_DEPLOY_ENV}--${_APP_NAME}", "--region", "us-west-2", "--cloudwatchregion", "us-west-1", "--workername", "MAGIC_ECS_TASK_ARN"]
