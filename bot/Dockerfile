FROM alpine:3.9
# 设置时区为上海
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add tzdata && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone
ADD bot /bot
# Application监听端口
# Manager监听端口
# Booker监听端口
# BotPlugin监听端口
EXPOSE 9091

ENTRYPOINT [ "/bot" ]