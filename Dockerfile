FROM scratch

COPY yaml-docs /usr/bin/

WORKDIR /yaml-docs

ENTRYPOINT ["yaml-docs"]
