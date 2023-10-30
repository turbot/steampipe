FROM debian:bullseye-slim
LABEL maintainer="Turbot Support <help@turbot.com>"

ARG TARGETVERSION
ARG TARGETARCH

# add a non-root 'steampipe' user
RUN adduser --system --disabled-login --ingroup 0 --gecos "steampipe user" --shell /bin/false --uid 9193 steampipe

# updates and installs - 'wget' for downloading steampipe, 'less' for paging in 'steampipe query' interactive mode
RUN apt-get update -y && apt-get install -y wget less && rm -rf /var/lib/apt/lists/*

# download the release as given in TARGETVERSION and TARGETARCH
RUN echo \
 && cd /tmp \
 && wget -nv https://github.com/turbot/steampipe/releases/download/${TARGETVERSION}/steampipe_linux_${TARGETARCH}.tar.gz \
 && tar xzf steampipe_linux_${TARGETARCH}.tar.gz \
 && mv steampipe /usr/local/bin/ \
 && rm -rf /tmp/steampipe_linux_${TARGETARCH}.tar.gz 

# Change user to non-root
USER steampipe:0

# Use a constant workspace directory that can be mounted to
WORKDIR /workspace

# disable auto-update
ENV STEAMPIPE_UPDATE_CHECK=false

# disable telemetry
ENV STEAMPIPE_TELEMETRY=none

# Create a temporary mod - this is required to make sure that the dashboard server starts without problems
RUN steampipe mod init

# Run steampipe service once
RUN steampipe service start --dashboard

# and stop it
RUN steampipe service stop

# Cleanup
# remove the generated service .passwd file from this image, so that it gets regenerated in the container
RUN rm -f /home/steampipe/.steampipe/internal/.passwd
# remove the temporary mod
RUN rm -f ./mod.sp

# expose postgres service default port
EXPOSE 9193

# expose dashboard service default port
EXPOSE 9194

COPY docker-entrypoint.sh /usr/local/bin
ENTRYPOINT [ "docker-entrypoint.sh" ]
CMD [ "steampipe"]
