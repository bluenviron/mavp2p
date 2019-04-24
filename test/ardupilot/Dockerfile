FROM amd64/debian:stretch-slim

# refs
# http://ardupilot.org/dev/docs/building-setup-linux.html#building-setup-linux
# https://github.com/ArduPilot/ardupilot/blob/master/Tools/environment_install/install-prereqs-ubuntu.sh
# http://ardupilot.org/dev/docs/setting-up-sitl-on-linux.html

RUN apt update && apt-get install -y --no-install-recommends \
    ca-certificates \
    git \
    python \
    python-future \
    python-setuptools \
    python-wheel \
    python-dev \
    python-pip \
    python-lxml \
    gcc \
    g++ \
    procps

RUN pip install pymavlink

RUN git clone https://github.com/ArduPilot/ardupilot \
    && cd ardupilot \
    && git checkout 3abe8fe \
    && git submodule update --init --recursive

ENTRYPOINT [ "sh", "-c", "cd ardupilot/APMrover2 \
    && ../Tools/autotest/sim_vehicle.py --no-mavproxy" ]
