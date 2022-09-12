FROM oraclelinux:8-slim
LABEL maintainer="Turbot Support <help@turbot.com>"

# to run tests from the branch
ARG TARGETBRANCH

# add a non-root 'steampipe' user
RUN adduser --system --shell /bin/false --uid 9193 --gid 0 --create-home steampipe

# updates and installs - 'wget' for downloading steampipe, 'less' for paging in 'steampipe query' 
# interactive mode, and others for running acceptance tests
RUN microdnf update -y && microdnf upgrade -y && microdnf install -y sudo findutils wget git jq sed vim curl bc tar less

# copy steampipe binary from local folder
COPY steampipe /usr/bin/

# Use a constant workspace directory that can be mounted to
WORKDIR /workspace

# change the owner of the /workspace directory
RUN chown steampipe /workspace

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

COPY run-tests.sh /usr/bin
ENTRYPOINT [ "sh", "-c", "run-tests.sh $BRANCH" ]
