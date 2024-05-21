# scripts/ubuntu/Dockerfile

FROM ubuntu:latest

# Install dependencies
RUN apt-get update && \
    apt-get install -y tar ca-certificates jq curl

# Create steampipe user
RUN useradd -m steampipe

# Copy scripts
COPY . /scripts
RUN chown -R steampipe:steampipe /scripts && chmod +x /scripts/*.sh

# Copy artifacts
COPY ../../artifacts /artifacts
RUN tar -xzf /artifacts/linux.tar.gz -C /usr/local/bin && \
    chown steampipe:steampipe /usr/local/bin/steampipe && \
    chmod +x /usr/local/bin/steampipe

# Set user to steampipe
USER steampipe

# Run smoke test script
ENTRYPOINT ["/scripts/smoke_test.sh"]