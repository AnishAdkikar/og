FROM ubuntu

WORKDIR /db

COPY og /db/

CMD ["./og"]