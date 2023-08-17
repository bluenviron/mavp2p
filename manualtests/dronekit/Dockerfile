FROM amd64/debian:stretch-slim

# refs
# http://python.dronekit.io/develop/sitl_setup.html

RUN apt update && apt-get install -y --no-install-recommends \
    ca-certificates \
    git \
    python-dev \
    python-setuptools \
    python-lxml \
    gcc \
    && rm -rf /var/lib/apt/lists/*

RUN git clone https://github.com/dronekit/dronekit-sitl \
    && cd /dronekit-sitl \
    && git checkout a63e97e \
    && python setup.py install \
    && rm -rf /dronekit-sitl

ENTRYPOINT [ "dronekit-sitl", "copter" ]
