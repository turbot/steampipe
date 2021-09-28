FROM debian:buster-slim
LABEL maintainer="Turbot Support <help@turbot.com>"

ARG TARGETVERSION
ARG TARGETOS
ARG TARGETARCH

#  'wget' for downloading steampipe, 'less' for paging in the UI
RUN apt-get update -y \
 && apt-get install -y wget less \
 && adduser --system --disabled-login --ingroup 0 --gecos "steampipe user" --shell /bin/false --uid 9193 steampipe

# downlaod the published image
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

# Run --version
RUN steampipe --version

# Run steampipe query to install db and fdw (they are installed on the first run)
RUN steampipe query "select * from steampipe_mod"

RUN rm -f /home/steampipe/.steampipe/internal/.passwd

EXPOSE 9193
COPY docker-entrypoint.sh /usr/local/bin
ENTRYPOINT [ "docker-entrypoint.sh" ]
CMD [ "steampipe"]
