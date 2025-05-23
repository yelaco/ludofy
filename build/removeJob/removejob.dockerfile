FROM amazonlinux:2

# Install necessary tools
RUN yum update -y && \
	yum install -y git python3 pip unzip wget tar && \
	curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" && \
	unzip awscliv2.zip && \
	./aws/install && \
	wget "https://github.com/aws/aws-sam-cli/releases/latest/download/aws-sam-cli-linux-x86_64.zip" && \
	unzip aws-sam-cli-linux-x86_64.zip -d sam-installation && \
	./sam-installation/install

# Set working directory
WORKDIR /app

# Copy deployment script from scripts folder into container
COPY scripts/remove.sh /remove.sh
RUN chmod +x /remove.sh

# Set the entrypoint to the deploy script
ENTRYPOINT ["/remove.sh"]
