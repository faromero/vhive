#!/bin/bash
sudo apt-get update

apt-get install -y btrfs-tools pkg-config libseccomp-dev unzip tar libseccomp2 socat util-linux apt-transport-https curl ipvsadm

wget -c https://github.com/google/protobuf/releases/download/v3.11.4/protoc-3.11.4-linux-x86_64.zip
sudo unzip protoc-3.11.4-linux-x86_64.zip -d /usr/local

wget https://golang.org/dl/go1.15.2.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.15.2.linux-amd64.tar.gz

export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile

export KUBECONFIG=/etc/kubernetes/admin.conf
echo 'export KUBECONFIG=/etc/kubernetes/admin.conf' >> /etc/profile

# Build and install runc and containerd
GOGITHUB=${HOME}/go/src/github.com/
RUNC_ROOT=${GOGITHUB}/opencontainers/runc
CONTAINERD_ROOT=${GOGITHUB}/containerd/containerd
mkdir -p $RUNC_ROOT
mkdir -p $CONTAINERD_ROOT

git clone https://github.com/opencontainers/runc.git $RUNC_ROOT
git clone -b cri_logging https://github.com/plamenmpetrov/containerd.git $CONTAINERD_ROOT

cd $RUNC_ROOT
make && make install

cd $CONTAINERD_ROOT
make && make install

containerd --version || echo "failed to build containerd"


# Install k8s
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
echo "deb http://apt.kubernetes.io/ kubernetes-xenial main" > /etc/apt/sources.list.d/kubernetes.list
apt update
apt install -y cri-tools ebtables ethtool kubeadm kubectl kubelet kubernetes-cni

# Use containerd for crictl
echo "runtime-endpoint: unix:///run/containerd/containerd.sock" > /etc/crictl.yaml

# Install knative CLI
git clone https://github.com/knative/client.git $HOME/client
cd $HOME/client
hack/build.sh -f
mv kn /usr/local/bin


# Necessary for containerd as container runtime but not docker
modprobe overlay
modprobe br_netfilter

# Set up required sysctl params, these persist across reboots.
cat <<EOF | sudo tee /etc/sysctl.d/99-kubernetes-cri.conf
net.bridge.bridge-nf-call-iptables  = 1
net.ipv4.ip_forward                 = 1
net.bridge.bridge-nf-call-ip6tables = 1
EOF

sudo sysctl --system
# ---------------------------------------------------------

swapoff -a
sysctl net.ipv4.ip_forward=1
