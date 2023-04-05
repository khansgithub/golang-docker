#!/bin/sh
apk add git
apk add openssh-server
# adduser --disabled-password git --gecos ""
adduser --disabled-password --shell "$(which git-shell)" git --gecos ""
echo "git:123" | chpasswd
mkdir /home/git/git-shell-commands
echo '#!/bin/sh' >> /home/git/git-shell-commands/git-init
echo "git init $@" >> /home/git/git-shell-commands/git-init
mkdir "/home/git/.ssh"
ssh-keygen -t rsa -N '' -b 4096 -f "/home/git/.ssh/id_rsa" 
cat "/home/git/.ssh/id_rsa.pub" >> "/home/git/.ssh/authorized_keys"
chown -R git:git /srv
ssh-keygen -A
echo 'PermitRootLogin no' >> /etc/ssh/sshd_config
echo 'StrictHostKeyChecking no' >> /etc/ssh/ssh_config
/usr/sbin/sshd

sleep infinitley

# git config --global user.email "you@example.com"
# git config --global user.name "Your Name"