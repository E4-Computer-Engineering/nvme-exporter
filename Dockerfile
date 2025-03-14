FROM fedora:41

RUN dnf update -y \
    && dnf install -y nvme-cli \
    && dnf clean all

COPY nvme_exporter /usr/bin/nvme_exporter

EXPOSE 9998
ENTRYPOINT ["/usr/bin/nvme_exporter"]
