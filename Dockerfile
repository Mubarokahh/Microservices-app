# Dockerfile for Redis
FROM redis:alpine

# Expose the default Redis port
EXPOSE 6379

# Start Redis
CMD ["redis-server"]