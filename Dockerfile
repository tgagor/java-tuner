FROM alpine:3

COPY bin/java-tuner /usr/local/bin/java-tuner

ENTRYPOINT [ "/usr/local/bin/java-tuner" ]
