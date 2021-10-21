FROM golang:1.16-buster

RUN apt-get -y update && apt-get -y install rsync build-essential bash zip unzip curl gettext jq python3-pip python3-dev git wget tar xz-utils
RUN pip3 install --upgrade pip

RUN wget "https://github.com/koalaman/shellcheck/releases/download/stable/shellcheck-stable.linux.x86_64.tar.xz" && \
    tar --xz -xvf shellcheck-stable.linux.x86_64.tar.xz && \
    cp shellcheck-stable/shellcheck /usr/bin/

RUN wget "https://dl.minio.io/server/minio/release/linux-amd64/minio" && \
    chmod +x minio && \
    mv minio /usr/bin/minio

# RUN wget "https://dl.minio.io/client/mc/release/linux-amd64/mc" && \
#     chmod +x mc && \
#     mv mc /usr/bin/mc

RUN wget https://releases.hashicorp.com/terraform/1.0.11/terraform_1.0.11_linux_amd64.zip && \
    unzip terraform_1.0.11_linux_amd64.zip -d /usr/bin

# install BOSH
RUN wget "https://github.com/cloudfoundry/bosh-cli/releases/download/v6.2.1/bosh-cli-6.2.1-linux-amd64" -O bosh
RUN chmod +x ./bosh && \
    mv ./bosh /usr/bin/bosh

# install credhub
RUN wget "https://github.com/cloudfoundry-incubator/credhub-cli/releases/download/2.6.2/credhub-linux-2.6.2.tgz" -O credhub.tgz
RUN tar xzf credhub.tgz
RUN chmod +x ./credhub && \
    mv ./credhub /usr/bin/credhub

# install fly
RUN wget "https://platform-automation.ci.cf-app.com/api/v1/cli?arch=amd64&platform=linux" -O fly
RUN chmod +x ./fly && \
    mv ./fly /usr/bin/fly

# install https://github.com/sclevine/yj
RUN wget "https://github.com/sclevine/yj/releases/download/v4.0.0/yj-linux" -O yj
RUN chmod +x ./yj && \
    mv ./yj /usr/bin/yj

ENV GO112MODULE=on

# install om
RUN git clone https://github.com/pivotal-cf/om
RUN cd om && \
    go build . && \
    mv om /usr/bin/ && \
    cd - && \
    rm -rf om

# rspec
RUN apt-get -y install ruby ruby-dev && \
    echo "gem: --no-document" >> /etc/gemrc && \
    gem install rspec english bundler

# uaac
RUN gem install cf-uaac

# used by `delete-terraformed-ops-manager`
RUN pip3 install awscli

# govc
RUN go get github.com/vmware/govmomi/govc

# openstack
RUN pip3 install python-openstackclient

# gcloud
RUN echo "deb http://packages.cloud.google.com/apt cloud-sdk-bionic main" | tee /etc/apt/sources.list.d/google-cloud-sdk.list
RUN curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
RUN apt-get -y update && apt-get -y install --no-install-recommends google-cloud-sdk

# azure
RUN echo "deb [arch=amd64] https://packages.microsoft.com/repos/azure-cli/ bionic main" | tee /etc/apt/sources.list.d/azure-cli.list
RUN curl -L https://packages.microsoft.com/keys/microsoft.asc | apt-key add -
RUN apt-get install apt-transport-https && apt-get update && apt-get install azure-cli


RUN git config --global user.name "platform-automation"
RUN git config --global user.email "platformautomation@groups.vmware.com"

# used by test
RUN go get github.com/onsi/ginkgo/ginkgo
RUN go get github.com/onsi/gomega

# use to cleanup IAASes
RUN wget -O /usr/bin/leftovers "https://github.com/genevieve/leftovers/releases/download/v0.59.0/leftovers-v0.59.0-linux-amd64"
RUN chmod +x /usr/bin/leftovers

ENV CGO_ENABLED=0

# used by act
RUN apt-get update && apt-get install -y \
    software-properties-common \
    npm
RUN npm install npm@latest -g && \
    npm install n -g && \
    n latest