FROM yikaus/alpine-bash 
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
ADD goch /
ENTRYPOINT /goch