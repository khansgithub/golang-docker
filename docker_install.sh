#!/bin/bash
curl -fsSL https://get.docker.com -o get-docker.sh
bash get-docker.sh && sudo usermod -aG docker $USER && newgrp docker
