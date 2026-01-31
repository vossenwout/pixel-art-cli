FROM golang:1.22-bookworm

ENV DEBIAN_FRONTEND=noninteractive \
    LANG=C.UTF-8 \
    LC_ALL=C.UTF-8 \
    TERM=xterm-256color

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates curl wget unzip xz-utils \
    git openssh-client gnupg \
    build-essential pkg-config \
    python3 python3-pip python3-venv \
    nodejs npm \
    ripgrep fd-find jq \
  && rm -rf /var/lib/apt/lists/*

RUN ln -s /usr/bin/fdfind /usr/local/bin/fd

WORKDIR /workspace

CMD ["bash","-lc","bash"]
