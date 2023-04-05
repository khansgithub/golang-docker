docker run \
	-d \
	--name git \
	-v "$(pwd)/"entrypoint.sh:/usr/home/entrypoint.sh \
    -v "$(pwd)/"repo:/srv/repo \
	--entrypoint "/usr/home/entrypoint.sh" \
	-p 2222:22 \
    --rm \
     --health-cmd "set -e;  nc -zv localhost 22; if [ $? -eq 0 ]; then exit 0; else exit 1; fi" \
     --health-interval=10s \
	alpine

ssh-keyscan -p 2222 localhost >> ~/.ssh/known_hosts
cd repo
git init
git add *
git config --local user.name "foo"
git config --local user.email "foo@bar.com"
git commit -m "initial commit"
mkdir ~/clone_dir
sudo ln -s ~/clone_dir clone_dir