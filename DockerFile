FROM alpine
COPY zwiftpower /
CMD ["/zwiftpower", "http"]