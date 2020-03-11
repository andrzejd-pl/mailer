FROM andrzejd/go-env

WORKDIR /go/src/mailer
COPY . ./
RUN go get -d ./...
RUN go install .

CMD ["mailer"]