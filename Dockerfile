From golang:latest


ADD . /go/src/github.com/ImamHyderAli/heartbeat


# Build the contact_registry command inside the container.


RUN go install github.com/ImamHyderAli/heartbeat


# Run the contact_registry command when the container starts.


ENTRYPOINT /go/bin/heartbeat


# http server listens on port 1234.


EXPOSE 1234
