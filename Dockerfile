FROM scratch
MAINTAINER Kelsey Hightower <kelsey.hightower@gmail.com>
ADD scheduler /scheduler
ENTRYPOINT ["/scheduler"]
