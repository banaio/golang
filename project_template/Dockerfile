FROM golang:1.13-alpine3.10

LABEL maintainer="Mohamed Bana <m@bana.io>"

ENV TERM xterm-256color
ENV GOOS linux
ENV GOARCH amd64
ENV CGO_ENABLED 0
ENV WORKDIR /app
ENV CMD_BINARY=project_template.exe
ENV CMD_BINARY_PATH=$WORKDIR/$CMD_BINARY

RUN \
  apk --no-cache update --no-cache && \
  apk --no-cache add --no-cache ca-certificates bash make git ncurses

# Add "--pretty-print", "-eufxv" to the list for additional debug output
SHELL ["/bin/bash", "--verbose", "-euf", "-o", "nounset", "-o", "errexit", "-o", "noclobber", "-o", "pipefail", "-c"]

WORKDIR $WORKDIR

# Build Cache #
# Copies go.mod go.sum to /app - was originally `COPY go.mod go.sum ./`.
COPY go.??? ./
# Now download the modules which should cache go modules to avoid re-fetching when rebuilding.
RUN go mod download
# Copy all the source files and then build them, this should result in quicker build times of the Docker image.
COPY . .
###############

# Build binary, then try to print help menu then add a sym link in `/usr/bin` to it to make life easier.
RUN \
  make build && \
  chmod +x ${CMD_BINARY_PATH} && \
  ${CMD_BINARY_PATH} --help && \
  echo "${CMD_BINARY_PATH}: binary runs" && \
  ln -s ${CMD_BINARY_PATH} /usr/bin/$CMD_BINARY && \
  chmod +x /usr/bin/$CMD_BINARY && \
  /usr/bin/$CMD_BINARY --help && \
  echo "/usr/bin/$CMD_BINARY: binary runs"

CMD ["/bin/bash", "-euf", "-o", "nounset", "-o", "errexit", "-o", "noclobber", "-o", "pipefail", "-c"]
ENTRYPOINT [ \
" \n\
RESET=`tput sgr0` \n\
RED=`tput bold; tput setaf 1` \n\
GREEN=`tput bold; tput setaf 2` \n\
printf -- '%s' ${GREEN} '---- started: ' 'time=' $(date --utc -Iseconds) ${RESET} $'\n' \n\
exit_code=0 \n\
/usr/bin/$CMD_BINARY || (exit_code=$? && printf -- '%s' ${RED} 'ERROR: ' 'exit_code=' ${exit_code} ${RESET} $'\n') \n\
printf -- '%s' ${GREEN} '---- ended: ' 'time=' $(date --utc -Iseconds) ', ' 'exit_code=' ${exit_code} ${RESET} $'\n' \n\
" \
]
