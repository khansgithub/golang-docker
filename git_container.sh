docker run \
	-d \
	--name git \
	-v "$(pwd)/"entrypoint.sh:/usr/home/entrypoint.sh \
	--entrypoint "/usr/home/entrypoint.sh" \
	-p 2222:22 \
    --rm \
     --health-cmd "set -e;  nc -zv localhost 22; if [ $? -eq 0 ]; then exit 0; else exit 1; fi" \
     --health-interval=10s \
	alpine
    