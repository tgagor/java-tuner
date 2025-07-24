FROM alpine:3

COPY bin/java-tuner /usr/local/bin/java-tuner

# fake java binary to avoid errors
RUN echo '#!/bin/sh' > /usr/local/bin/java && \
    echo 'echo $(basename "$0") $@' >> /usr/local/bin/java && \
    chmod +x /usr/local/bin/java

ENTRYPOINT [ "/usr/local/bin/java-tuner" ]
