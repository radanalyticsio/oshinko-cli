FROM centos:latest

RUN yum install -y epel-release && yum clean all
RUN yum install -y python-pip && yum clean all

RUN pip install requests

COPY elastic_worker_daemon.py /opt

RUN chmod a+rx /opt/elastic_worker_daemon.py

CMD ["/opt/elastic_worker_daemon.py"]
