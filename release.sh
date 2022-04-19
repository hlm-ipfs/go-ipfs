# declare
export GROUP_NAME=ipfs2021
export PROJECT_NAME=ipfs

# build
GOOS=linux go build  -o ./bin/$PROJECT_NAME ./cmd/ipfs
# docker
IMAGE_TAG=registry.cn-hangzhou.aliyuncs.com/$GROUP_NAME/$PROJECT_NAME:latest
docker build -t $IMAGE_TAG --build-arg ARG_PROJECT_NAME=$PROJECT_NAME --build-arg ARG_CI_BUILD_INFO="$(date "+%Y-%m-%d %H:%M:%S")" .
docker push $IMAGE_TAG
docker rmi $IMAGE_TAG
docker image prune -f

# clean
rm -rf ./bin/$PROJECT_NAME
