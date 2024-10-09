FROM gcc:latest
WORKDIR /app
COPY app.c /app/
RUN gcc -o server app.c
CMD ["./server"]
