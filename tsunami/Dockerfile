FROM ubuntu:18.04 AS builder

RUN apt-get update -y && apt-get upgrade -y

RUN apt-get install -y git openjdk-11-jdk && rm -rf /var/lib/apt/lists/*

ADD https://api.github.com/repos/google/tsunami-security-scanner/compare/master...HEAD /dev/null
RUN git clone https://github.com/google/tsunami-security-scanner /opt/tsunami-security-scanner
ADD https://api.github.com/repos/google/tsunami-security-scanner-plugins/compare/master...HEAD /dev/null
RUN git clone https://github.com/google/tsunami-security-scanner-plugins /opt/tsunami-security-scanner-plugins

WORKDIR /opt/tsunami-security-scanner
RUN ./gradlew shadowJar

WORKDIR /opt/tsunami-security-scanner-plugins/google
RUN ./build_all.sh

FROM ubuntu:18.04

RUN apt-get update -y && apt-get upgrade -y

RUN apt-get install -y openjdk-11-jdk nmap ncrack && rm -rf /var/lib/apt/lists/*

COPY --from=builder /opt/tsunami-security-scanner-plugins/google/build/plugins /opt/tsunami/plugins
COPY --from=builder /opt/tsunami-security-scanner/main/build/libs/tsunami-main-*-cli.jar /opt/tsunami/tsunami-main-cli.jar
COPY --from=builder /opt/tsunami-security-scanner/tsunami.yaml /opt/tsunami/tsunami.yaml

WORKDIR /opt/tsunami
ENTRYPOINT ["java", "-cp", "tsunami-main-cli.jar:/opt/tsunami/plugins/*", "-Dtsunami-config.location=/opt/tsunami/tsunami.yaml", "com.google.tsunami.main.cli.TsunamiCli", "--scan-results-local-output-format=JSON", "--scan-results-local-output-filename=/dev/stdout"]
CMD ["--ip-v4-target=127.0.0.1"]
