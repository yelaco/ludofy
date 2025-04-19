FROM amazonlinux:2

# Install necessary tools
RUN yum update -y && \
  yum install -y git python3 pip unzip wget && \
  wget https://go.dev/dl/go1.24.1.linux-amd64.tar.gz && \
  rm -rf /usr/local/go && tar -C /usr/local -xzf go1.24.1.linux-amd64.tar.gz && \
  curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" && \
  unzip awscliv2.zip && \
  ./aws/install && \
  wget "https://github.com/aws/aws-sam-cli/releases/latest/download/aws-sam-cli-linux-x86_64.zip" && \
  unzip aws-sam-cli-linux-x86_64.zip -d sam-installation && \
  ./sam-installation/install

# Set working directory
WORKDIR /app

# Copy deployment script from scripts folder into container
COPY scripts/deploy.sh /deploy.sh
RUN chmod +x /deploy.sh

# Set the entrypoint to the deploy script
ENTRYPOINT ["/deploy.sh"]
