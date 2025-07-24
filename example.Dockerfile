FROM amazoncorretto:17

ADD https://github.com/tgagor/java-tuner/releases/latest/download/java-tuner-linux-amd64 /usr/local/bin/java-tuner
RUN chmod +x /usr/local/bin/java-tuner
ENTRYPOINT ["java-tuner", "--log-format", "plain", "--"]

# COPY my-app.jar ./
CMD ["-XX:+PrintCommandLineFlags", "-version"]
