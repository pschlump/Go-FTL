
#go build -o corp-server 2>&1 | color-cat -c red
#CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o $APP_NAME ./ && \

GOOS=linux go build -o corp-server.linux 

