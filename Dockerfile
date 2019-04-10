FROM alpine:3.5

EXPOSE 9090

COPY ./octopus /octopus

# ENTRYPOINT [ "/octopus" ]