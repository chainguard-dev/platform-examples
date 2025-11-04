ARG ORG_NAME
ARG BASE_IMAGE_NAME=chainguard-base:latest
FROM cgr.dev/${ORG_NAME}/${BASE_IMAGE_NAME}

RUN apk add --no-cache aws-cli bash chainctl crane jq curl

COPY image-copy.sh /image-copy.sh

ENTRYPOINT ["bash", "/image-copy.sh"]
