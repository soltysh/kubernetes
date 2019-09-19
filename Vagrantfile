# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure(2) do |config|
  config.vm.box = "bento/ubuntu-20.04"
  config.vm.provider "virtualbox" do |v|
    v.cpus = 4
    v.memory = 8192
  end

  config.vm.network "private_network", type: "dhcp"
  config.vm.synced_folder ".", "/go/src/k8s.io/kubernetes/", type: "nfs"

  config.vm.provision "file", source: "~/.gitconfig", destination: ".gitconfig"
  config.vm.provision "shell", inline: <<-SHELL
    set -e -x -u
    apt-get update -y || (sleep 40 && apt-get update -y)
    apt-get install -y git gcc-multilib gcc-mingw-w64 jq git-secrets
    wget -qO- https://storage.googleapis.com/golang/go1.15.linux-amd64.tar.gz | tar -C /usr/local -xz
    echo 'export GOPATH=/go' >> /root/.bashrc
    echo 'ulimit -n 65535' >> /root/.bashrc
    echo 'export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin' >> /root/.bashrc
    cp .gitconfig /root/
  SHELL
end
