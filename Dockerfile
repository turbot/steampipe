FROM debian:bullseye-slim
LABEL maintainer="Turbot Support <help@turbot.com>"

ARG TARGETVERSION
ARG TARGETOS
ARG TARGETARCH

# add a non-root 'steampipe' user
RUN adduser --system --disabled-login --ingroup 0 --gecos "steampipe user" --shell /bin/false --uid 9193 steampipe

# updates and installs - 'wget' for downloading steampipe, 'less' for paging in 'steampipe query' interactive mode
RUN apt-get update -y && apt-get install -y wget less

# download the release as given in TARGETVERSION, TARGETOS and TARGETARCH
RUN echo \
 && cd /tmp \
 && wget -nv https://github.com/turbot/steampipe/releases/download/${TARGETVERSION}/steampipe_${TARGETOS}_${TARGETARCH}.tar.gz \
 && tar xzf steampipe_${TARGETOS}_${TARGETARCH}.tar.gz \
 && mv steampipe /usr/local/bin/ \
 && rm -rf /tmp/steampipe_${TARGETOS}_${TARGETARCH}.tar.gz 

# Change user to non-root
USER steampipe:0

# Use a constant workspace directory that can be mounted to
WORKDIR /workspace

# disable auto-update
ENV STEAMPIPE_UPDATE_CHECK=false

# disable telemetry
ENV STEAMPIPE_TELEMETRY=none

# Run steampipe service once
RUN steampipe service start --dashboard
# and stop it
RUN steampipe service stop

# remove the generated service .passwd file from this image, so that it gets regenerated in the container
RUN rm -f /home/steampipe/.steampipe/internal/.passwd

# expose postgres service default port
EXPOSE 9193

# expose dashboard service default port
EXPOSE 9194

COPY docker-entrypoint.sh /usr/local/bin
ENTRYPOINT [ "docker-entrypoint.sh" ]
CMD [ "steampipe"]
