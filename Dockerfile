FROM alpine:3.20

COPY yaml-docs /usr/bin/

WORKDIR /yaml-docs

ENTRYPOINT ["yaml-docs"]
