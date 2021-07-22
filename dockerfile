FROM debian:buster-slim
LABEL maintainer="Turbot Support <help@turbot.com>"

ENV STEAMPIPE_VERSION  0.6.0
ARG TARGETOS # OS component of TARGETPLATFORM
ARG TARGETARCH # architecture component of TARGETPLATFORM
USER root:0

RUN apt-get update -y
RUN adduser --system --disabled-login --ingroup 0 --gecos "steampipe user" --shell /bin/false --uid 9193 steampipe

# need wget just for the install download
# need 'less' for paging in the UI
RUN apt-get install -y wget less

# Change user to non-root
USER steampipe:0

# Use a constant workspace directory that can be mounted to
WORKDIR /workspace

# commented out for testing...
# RUN echo \
#  && cd /tmp \
#  && wget -nv https://github.com/turbot/steampipe/releases/download/${STEAMPIPE_VERSION}/steampipe_${TARGETOS}_${TARGETARCH}.tar.gz \
#  && tar xzf steampipe_${TARGETOS}_${TARGETARCH}.tar.gz \
#  && mv steampipe /usr/local/bin/ \
#  && rm -rf /tmp/steampipe_${TARGETOS}_${TARGETARCH}.tar.gz 
# added for testing....


COPY steampipe /usr/local/bin/
ENV STEAMPIPE_UPDATE_CHECK=false

# Install the steampipe plugin.  This will also install db and fdw,
# as they are installed on the first run
RUN steampipe plugin install steampipe
EXPOSE 9193
COPY docker-entrypoint.sh /usr/local/bin
ENTRYPOINT [ "docker-entrypoint.sh" ]
CMD [ "steampipe", "query" ]