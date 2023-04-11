FROM debian:buster

# Install wget, lsb-release and curl
RUN apt update && \
    apt install -y wget lsb-release curl fuse libfuse3-dev tree ca-certificates e2fsprogs gnupg2 && \
    # security updates
    apt install -y apt systemd

# Add key
RUN wget -O - http://repository.moosefs.com/moosefs.key | apt-key add -
RUN echo "deb http://repository.moosefs.com/4.44.4/apt/debian/buster buster main" > /etc/apt/sources.list.d/moosefs.list

# Install oras
# ENV ORAS_VERSION="1.0.0"
# RUN curl -LO "https://github.com/oras-project/oras/releases/download/v${VERSION}/oras_${VERSION}_linux_amd64.tar.gz" && \
#    mkdir -p oras-install/ && \
#    tar -zxf oras_${VERSION}_*.tar.gz -C oras-install/ && \
#    mv oras-install/oras /usr/local/bin/ && \
#    rm -rf oras_${VERSION}_*.tar.gz oras-install/

# Copy the CSI plugin
ADD oras-csi-plugin /bin/

ENTRYPOINT ["/bin/oras-csi-plugin"]
