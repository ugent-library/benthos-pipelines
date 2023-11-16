FROM jeffail/benthos:4.21

COPY /mongo-projects-to-projects/config.yaml .

ENTRYPOINT ["/benthos"]

CMD ["-c", "/config.yaml"]