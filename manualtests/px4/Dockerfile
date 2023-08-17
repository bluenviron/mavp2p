FROM amd64/debian:stretch-slim

# refs
# https://dev.px4.io/en/simulation/
# https://raw.githubusercontent.com/PX4/Devguide/master/build_scripts/ubuntu_sim_common_deps.sh
# https://raw.githubusercontent.com/PX4/Devguide/master/build_scripts/ubuntu_sim.sh
# https://github.com/PX4/Firmware

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    git \
    make \
    cmake \
    binutils \
    g++ \
    unzip \
    python-pip \
    python-empy \
    python-toml \
    python-numpy \
    python-yaml \
    && rm -rf /var/lib/apt/lists/*

RUN pip install jinja2

RUN echo "deb http://deb.debian.org/debian buster main" >> /etc/apt/sources.list \
    && apt-get update && apt-get install -y --no-install-recommends -t buster \
    gazebo9 libgazebo9-dev libopencv-dev protobuf-compiler libeigen3-dev

RUN git clone -b v1.9.0-beta1 https://github.com/PX4/Firmware px4

WORKDIR /px4

# enable udp broadcasting
RUN sed -i '/param set WEST_EN 0/a param set MAV_BROADCAST 1' ./ROMFS/px4fmu_common/init.d-posix/rcS

# set mavlink version
# RUN sed -i '/param set WEST_EN 0/a param set MAV_PROTO_VER 1' ./ROMFS/px4fmu_common/init.d-posix/rcS

RUN make px4_sitl

ENV HEADLESS 1

ENTRYPOINT [ "make", "px4_sitl", "gazebo" ]
