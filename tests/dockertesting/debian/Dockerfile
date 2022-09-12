FROM debian:bullseye-slim
LABEL maintainer="Turbot Support <help@turbot.com>"

# to run tests from the branch
ARG TARGETBRANCH

# add a non-root 'steampipe' user
RUN adduser --system --disabled-login --ingroup 0 --gecos "steampipe user" --shell /bin/false --uid 9193 steampipe

# updates and installs - 'wget' for downloading steampipe, 'less' for paging in 'steampipe query' interactive mode,
# and others for running acceptance tests
RUN apt-get update -y && apt-get install -y sudo wget git jq sed vim curl bc less

# copy steampipe binary from local folder
COPY steampipe /usr/local/bin/

# Use a constant workspace directory that can be mounted to
WORKDIR /workspace

# change the owner of the /workspace directory
RUN chown steampipe:0 /workspace

# Change user to non-root
USER steampipe:0

# disable auto-update
ENV STEAMPIPE_UPDATE_CHECK=false

# disable telemetry
ENV STEAMPIPE_TELEMETRY=none

# enable introspection tables
ENV STEAMPIPE_INTROSPECTION=info

# use to run tests from the branch
ENV BRANCH=$TARGETBRANCH

# expose postgres service default port
EXPOSE 9193

# expose dashboard service default port
EXPOSE 9194

COPY run-tests.sh /usr/local/bin
ENTRYPOINT [ "sh", "-c", "run-tests.sh $BRANCH" ]
