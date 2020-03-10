FROM andrzejd/go-env

RUN go get -u github.com/andrzejd-pl/mailer
RUN go install github.com/andrzejd-pl/mailer

CMD ["mailer"]