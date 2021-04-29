FROM ubuntu:20.04

RUN apt-get update && \
    apt-get install -y curl && \
    useradd -ms /bin/bash steampipe

COPY $GITHUB_WORKSPACE / 

RUN tar -xf /linux.tar.gz -C /usr/local/bin

USER steampipe

WORKDIR /home/steampipe

RUN steampipe plugin install steampipe

ENTRYPOINT [ "/usr/local/bin/steampipe" ]