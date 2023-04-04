docker run --rm --name="test" -v "$(pwd)/foo.py:/usr/home/foo.py" -w /usr/home python:latest python foo.py

docker run --name="not-foo" -v "$(pwd)/foo.py:/usr/home/foo.py" -w /usr/home python:latest python -u foo.py


# docker run -it --name="test" -v "$(pwd)/foo.py:/usr/home/foo.py" python:latest "/bin/bash"

# docker run -it --rm --name test -v "$PWD":/usr/src/myapp -w /usr/src/myapp python:3 python foo.py
