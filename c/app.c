#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <netinet/in.h>

#define PORT 8000

int main() {
    int server_fd, new_socket;
    long valread;
    struct sockaddr_in address;
    int addrlen = sizeof(address);

    char *response = "HTTP/1.1 200 OK\r\nContent-Length: 21\r\nContent-Type: text/plain\r\n\r\nHello, World from C!\n";

    if ((server_fd = socket(AF_INET, SOCK_STREAM, 0)) == 0) {
        perror("In socket");
        exit(EXIT_FAILURE);
    }

    int opt = 1;
    if (setsockopt(server_fd, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt))) {
        perror("In setsockopt");
        exit(EXIT_FAILURE);
    }

    address.sin_family = AF_INET;
    address.sin_addr.s_addr = INADDR_ANY;
    address.sin_port = htons(PORT);

    if (bind(server_fd, (struct sockaddr *)&address, sizeof(address))<0) {
        perror("In bind");
        exit(EXIT_FAILURE);
    }
    if (listen(server_fd, 10) < 0) {
        perror("In listen");
        exit(EXIT_FAILURE);
    }
    printf("Server started at http://localhost:%d\n", PORT);
    while(1) {
        if ((new_socket = accept(server_fd, (struct sockaddr *)&address, (socklen_t*)&addrlen))<0) {
            perror("In accept");
            exit(EXIT_FAILURE);
        }
        char buffer[30000] = {0};
        valread = read(new_socket , buffer, 30000);

        write(new_socket , response , strlen(response));
        close(new_socket);
    }
    return 0;
}
