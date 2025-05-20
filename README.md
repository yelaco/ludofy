# 🎮 Ludofy – Serverless Game Backend Deployment Platform (PaaS)

Ludofy is a **Platform-as-a-Service (PaaS)** solution designed to simplify the deployment and management of backend services for online games. Built entirely on **AWS serverless architecture**, Ludofy enables game developers to quickly deploy scalable, event-driven, and cost-efficient game backends—without managing infrastructure.

https://ludofy.vercel.app/
## 🚀 Key Features

- 🧩 **Modular Backend Services** – Built-in support for:
  - Matchmaking
  - Player ranking (ELO, Glicko, TrueSkill)
  - Live chat & messaging
  - Friends & social systems
  - Match spectating
- ⚙️ **Serverless by Design** – Powered by AWS Lambda, API Gateway, DynamoDB, AppSync, SQS, SNS, and Fargate.
- 📦 **Custom Game Server Deployment** – Upload your Docker image, and Ludofy handles orchestration, scaling, and lifecycle management via ECS Fargate.
- 📈 **Real-time Monitoring** – Frontend dashboard to track active matches, CPU/memory usage, and deployment state.
- 💻 **SDK & API Integration** – Use the Go SDK to integrate your game logic, or interact directly with provided APIs.
- 🧠 **Auto-scaling Game Servers** – Servers dynamically start or stop based on active players and match load.

## 🧱 System Architecture

Ludofy consists of:
- **Frontend** (Vue + TailwindCSS): Platform dashboard to configure and deploy backends.
- **Deployment Engine** (AWS SAM, CloudFormation): Automates backend provisioning.
- **Game Backend Services**: Serverless microservices implementing core features.
- **ECS Fargate Game Server Layer**: Stateless, isolated microservers per match.

## 🛠️ Tech Stack

| Layer               | Tools / Services                        |
|--------------------|------------------------------------------|
| Infrastructure      | AWS SAM, CloudFormation, IaC            |
| Compute             | AWS Lambda, ECS Fargate                 |
| Communication       | API Gateway (HTTP/WebSocket), AppSync  |
| Storage & DB        | DynamoDB, S3                            |
| Messaging           | SQS, SNS                                |
| Auth                | Cognito                                 |
| Monitoring          | CloudWatch                              |
| Frontend            | Vue.js + TailwindCSS                    |
| Load Testing        | k6                                      |
| CI/CD               | GitHub Actions                          |

## 🧪 Performance & Scalability

- Tested with up to **10,000 concurrent players**
- Stable matchmaking latency under heavy load
- Dynamic scaling of game servers based on match requests
- Cost-efficient (serverless pay-per-use model)

## 🔧 Prerequisites

Before setting up Ludofy locally or deploying your backend, make sure the following tools are installed:

| Tool | Description | Install Guide |
|------|-------------|----------------|
| [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html) | Required for deploying backend via AWS SAM | Run `aws configure` to set credentials |
| [Docker](https://docs.docker.com/get-docker/) | Needed for building and uploading custom game servers | Used for ECS Fargate container deployment |
| [Task](https://taskfile.dev/#/installation) | Task runner used to simplify local dev workflows | Runs scripts like `task web:run-dev` |
| [Go](https://golang.org/dl/) | For SDK usage or backend service development | Required if customizing backend logic |
| [Node.js & npm](https://nodejs.org/) | Needed for running the frontend dashboard | Vue.js-based UI setup |
| [SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html) | AWS Serverless Application Model CLI | Used for deploying infrastructure via `task stack:deploy` |

## 🧰 Getting Started

You can try the production version directly:  
👉 **[https://ludofy.vercel.app](https://ludofy.vercel.app)**

Or clone and run locally:

1. **Clone the repo**  
   ```bash
   git clone https://github.com/yelaco/ludofy.git
   cd ludofy
   ```

2. **Frontend setup**
   ```bash
   task env:base
   task env:web

   cd web
   npm install
   task web:run-dev
   ```

4. **Deploy backend (via AWS SAM)**  
   ```bash
   task env:base
   task stack:deploy
   ```

## 📚 Documentation

- [Usage Guide](https://yelaco/ludofy)
- [SDK Reference](https://github.com/yelaco/ludofy)
- [Customization Templates](https://ludofy.vercel.app/help/customization)

## 📊 Evaluation

| Metric               | Result                     |
|----------------------|----------------------------|
| Max Users Tested     | 10,000 concurrent          |
| Avg Matchmaking Time | < 30 seconds               |
| Avg Server Latency   | < 100ms                    |
| Cost Efficiency      | ~70% cheaper than EC2/EKS  |

## 📜 License

This project is under the MIT License. See `LICENSE` for more details.

---

> 🧑‍🎓 Developed as a graduation project at Vietnam National University – University of Engineering and Technology, under the supervision of Dr. Phạm Mạnh Linh.
