FROM alpine:latest as zigbuild

RUN apk update && \
    apk add \
        curl \
        xz \
        tar

ARG ZIGTAR=zig-linux-x86_64-0.12.0-dev.2285+b80cad248.tar.xz

RUN mkdir -p /deps/zig

RUN curl -L https://ziglang.org/builds/$ZIGTAR | \
    tar -xJ --strip-components=1 -C /deps/zig/

FROM crazymax/osxcross:latest-alpine AS osxcross

FROM --platform=linux/amd64 golang:1.21

COPY /zigtool /zigtool
RUN cd /zigtool/zigcc && go install
RUN cd /zigtool/zigcpp && go install

COPY --from=zigbuild /deps/zig/ /deps/zig/

COPY --from=osxcross /osxcross/SDK/MacOSX13.1.sdk /osxsdk

ENV PATH="${PATH}:/deps/zig/"
ENV CGO_CPPFLAGS="-Wno-error -Wno-nullability-completeness -Wno-expansion-to-defined -Wdeprecated-declarations"

# MACOS
ENV OSXSDK="/osxsdk"
ENV CC="zigcc --sysroot=${OSXSDK} -I${OSXSDK}/usr/include -L${OSXSDK}/usr/lib -F${OSXSDK}/System/Library/Frameworks"
ENV CXX="zigcpp --sysroot=${OSXSDK} -I${OSXSDK}/usr/include -L${OSXSDK}/usr/lib -F${OSXSDK}/System/Library/Frameworks"
ENV CGO_CPPFLAGS="-w"
#  -framework Cocoa
# LINUX
#ENV CC="zigcc"
#ENV CXX="zigcpp"