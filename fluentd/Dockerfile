FROM fluent/fluentd-kubernetes-daemonset:v1.4-debian-elasticsearch
ARG GIT_SHA
RUN if [ -z "$GIT_SHA" ]; then echo "GIT_SHA not set"; exit 1; else : ; fi

RUN gem install grpc:1.20.0
RUN gem install memory_profiler

COPY ./lib/envelope_pb.rb /usr/local/lib/ruby/site_ruby/
COPY ./lib/ingress_pb.rb /usr/local/lib/ruby/site_ruby/
COPY ./lib/ingress_services_pb.rb /usr/local/lib/ruby/site_ruby/
COPY ./foo.rb /fluentd/vendor/bundle/ruby/2.6.0/gems/fluentd-1.4.2/bin/fluentd

COPY ./plugins/loggregator.rb /fluentd/plugins/out_loggregator.rb
LABEL org.opencontainers.image.revision=$GIT_SHA \
      org.opencontainers.image.source=https://code.cloudfoundry.org/eirini
