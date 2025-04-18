FROM amazonlinux:2

# Install necessary tools
RUN yum update -y && \
  yum install -y git python3 pip unzip && \
  curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" && \
  unzip awscliv2.zip && \
  ./aws/install && \
  pip3 install aws-sam-cli

# Set working directory
WORKDIR /app

# Copy deployment script from scripts folder into container
COPY scripts/deploy.sh /deploy.sh
RUN chmod +x /deploy.sh

# Set the entrypoint to the deploy script
ENTRYPOINT ["/deploy.sh"]
