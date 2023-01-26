#!/usr/bin/env bash

LOCAL_BIN="/usr/local/bin"

DEFAULT_ARCH="amd64"
DEFAULT_OS="linux"

#Terraform
DEFAULT_TERRAFORM_VERSION="1.3.5"

#Kubectl
DEFAULT_KUBECTL_VERSION="v1.26.1"

#Terragrunt
DEFAULT_TERRAGRUNT_VERSION="v0.42.7"

#Golang
DEFAULT_GOLANG_VERSION="1.19.5"

function download() {
    FROM=$1
    TO=$2
    curl --silent --show-error --fail -L -o $TO $FROM
}

function log_context() {
  NAME=$1
  VERSION=$2
  DEST_PATH=$3
  ARCH=$4
  LINK=$5
  echo "Installing ${NAME} version: ${VERSION} in ${DEST_PATH}, arch ${ARCH}, from $LINK"
}

function install_kubectl() {
    ARCH="${1:-$DEFAULT_ARCH}"
    VERSION="${2:-$DEFAULT_KUBECTL_VERSION}"
    LINK="https://dl.k8s.io/release/${VERSION}/bin/linux/${ARCH}/kubectl"

    log_context "terraform" $VERSION $LOCAL_BIN $ARCH $LINK
    download $LINK  $LOCAL_BIN/kubectl && \
    chmod +x $LOCAL_BIN/kubectl
}

function install_terraform() {
    ARCH="${1:-$DEFAULT_ARCH}"
    VERSION="${2:-$DEFAULT_TERRAFORM_VERSION}"
    LINK="https://releases.hashicorp.com/terraform/${VERSION}/terraform_${VERSION}_linux_${ARCH}.zip"

    log_context "terraform" $VERSION $LOCAL_BIN $ARCH $LINK
    download $LINK /tmp/terraform.zip && \
    unzip -o /tmp/terraform.zip -d /usr/local/bin && rm -f /tmp/terraform.zip
}

function install_terragrunt() {
    ARCH=${1:-$DEFAULT_ARCH}
    VERSION=${2:-$DEFAULT_TERRAGRUNT_VERSION}
    LOCAL_BIN="/usr/local/bin"
    LINK=https://github.com/gruntwork-io/terragrunt/releases/download/${VERSION}/terragrunt_linux_${ARCH}

    log_context "terragrunt" $VERSION $LOCAL_BIN $ARCH $LINK
    download $LINK $LOCAL_BIN/terragrunt
    chmod +x $LOCAL_BIN/terragrunt
}

function install_golang() {
    ARCH=${1:-$DEFAULT_ARCH}
    VERSION=${2:-$DEFAULT_GOLANG_VERSION}
    GOOS=${3:-$DEFAULT_OS}
    GOPATH=$HOME/go
    GOROOT="/usr/local/go"
    GOBIN=$GOPATH/bin
    mkdir -p $GOPATH
    LINK="https://dl.google.com/go/go${VERSION}.$GOOS-$ARCH.tar.gz"

    log_context "golang" $VERSION $LOCAL_BIN $ARCH $LINK
    download $LINK /tmp/go.tar.gz
    rm -rf /usr/local/go && tar -C /usr/local -xzf /tmp/go.tar.gz && mkdir -p $GOBIN

cat <<EOF > /etc/profile.d/golang.sh
      export GOROOT=$GOROOT
      export PATH=$PATH:/usr/local/go/bin:$GOBIN
EOF
}

function install_docker_engine() {
    mkdir -p /etc/apt/keyrings
    rm -f /etc/apt/keyrings/docker.gpg
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    echo \
      "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
      $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

    apt-get update
    chmod a+r /etc/apt/keyrings/docker.gpg
    apt-get update
    apt-get -y install docker-ce docker-ce-cli containerd.io docker-compose-plugin
    systemctl enable --now docker
}

function install_all_tools() {
    install_docker_engine
    install_golang
    install_kubectl
    install_terraform
    install_terragrunt

}