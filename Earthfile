VERSION 0.8
FROM golang:1.23

ARG --global registryName=
ARG --global imageName=

pre-build:
    ARG useVendorFolder=false
    WORKDIR /project
    # RUN git config --global url."ssh://git@github.com/<PRIVATE REPO>/".insteadOf https://github.com/<PRIVATE REPO>/
    # RUN mkdir -p -m 0600 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts
    COPY . .
    IF [ "$useVendorFolder" = "false" ]
        # RUN --ssh GOPRIVATE="github.com/<PRIVATE REPO>/*" go mod vendor
        RUN --ssh go mod vendor
    END
    SAVE ARTIFACT /project

build:
    FROM DOCKERFILE -f Dockerfile +pre-build/project/*

build-image:
    FROM +build
    SAVE IMAGE $registryName/$imageName:latest

push-image:
    ARG imageTag=
    FROM +build-image
    SAVE IMAGE --push $registryName/$imageName:$imageTag

deploy-test:
    BUILD +push-image --imageTag=main

deploy-prod:
    BUILD +push-image --imageTag=release
