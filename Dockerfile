# Use the latest Alpine Linux image
FROM alpine:latest

# Install OpenSSH
RUN apk update && \
    apk add --no-cache openssh

# Create a new user "testuser" with a home directory and set its password to "testpass"
RUN adduser -D -h /home/testuser testuser && \
    echo "testuser:testpass" | chpasswd

# Ensure /home/testuser exists and has the correct ownership
RUN mkdir -p /home/testuser && \
    chown testuser:testuser /home/testuser

# Generate SSH host keys (ensures keys are available at runtime)
RUN ssh-keygen -A

# Update SSH configuration:
# - Enable password authentication
# - Disable root login for security
RUN sed -i 's/#PasswordAuthentication.*/PasswordAuthentication yes/' /etc/ssh/sshd_config && \
    sed -i 's/#PermitRootLogin.*/PermitRootLogin no/' /etc/ssh/sshd_config

# Expose the SSH port
EXPOSE 22

# Start the SSH daemon in the foreground
CMD ["/usr/sbin/sshd", "-D"]
