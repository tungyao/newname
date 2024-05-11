FROM ubuntu
COPY . /app
WORKDIR /app
RUN chmod +x /app/newname
CMD ["/app/newname"]