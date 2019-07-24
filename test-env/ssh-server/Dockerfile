FROM alpine:3.6

RUN apk update && apk add \
      shadow \
      libcap \
      openssh \
      tcpdump \
      supervisor

COPY sshd_config /etc/ssh/sshd_config
COPY motd /etc/motd
RUN /usr/bin/ssh-keygen -A

RUN addgroup -S mole && adduser -S mole -G mole -D -s /bin/ash && usermod -p 'this-is-not-a-valid-hash' mole
RUN mkdir -p /home/mole/.ssh && chown mole:mole /home/mole/.ssh

RUN chgrp mole /usr/sbin/tcpdump && chmod 750 /usr/sbin/tcpdump && setcap cap_net_raw+ep /usr/sbin/tcpdump

COPY authorized_keys /home/mole/.ssh/

COPY supervisord.conf /etc/supervisord.conf
RUN mkdir -p /var/log/supervisor

#CMD /usr/sbin/sshd -D
ENTRYPOINT ["/usr/bin/supervisord"]
CMD ["-n", "-c", "/etc/supervisord.conf"]
