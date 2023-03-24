FROM golang:1.19 AS builder
WORKDIR /app/
ADD . /app
RUN make

FROM golang:1.19
WORKDIR /root/
COPY --from=builder /app/delta-dm ./
CMD ./delta daemon
EXPOSE 1414