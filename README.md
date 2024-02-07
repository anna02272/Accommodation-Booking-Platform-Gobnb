# Accommodation Booking Platform - Gobnb

This repository contains the implementation of a platform for offering and booking accommodations. This project is part of the "Service-Oriented Architectures and NoSQL Databases" course.

### Launch Guide:

Below are the steps to get the project up and running on your local environment.

#### Prerequisites:
1. Docker
2. Go programming language
3. Goland IDE
4. Minikube
5. Kubectl
6. Node.js
7. Visual Studio Code

#### Step 1: Clone the Repository
Clone the Gobnb repository to your local machine:

```bash
git clone https://github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024
```

#### Step 2: Import the Project in Goland
Open Goland IDE and import the cloned repository.

#### Step 3: Run Go Mod Tidy
Ensure all necessary dependencies are downloaded:

```bash
go mod tidy
```

#### Step 4: Build Docker Images and Start Docker Compose (or Kubernetes)
Option 1: Using Docker Compose:
```bash
docker-compose build
docker-compose up
```

Option 2: Using Minikube and Kubectl:
```bash
docker-compose build
minikube start
minikube load images <builded_image_name>
cd SOA_NoSQL_IB-MRS-2023-2024/Airbnb_backend
kubectl apply -f <file.yml> # Apply this for all .yml files in the directory
minikube tunnel
```

#### Step 5: Start Angular Frontend
1. Open Visual Studio Code.
2. Navigate to the `Airbnb_frontend` directory in the cloned repository.
3. Start the Angular app with npm:

```bash
npm install
npm start
```

#### Step 6: Access the Platform
Once the services are up and running, you can access the platform via the provided endpoints:

- **Frontend:** Open a web browser and go to [https://localhost:4200/home](https://localhost:4200/home)

#### Note:
- Ensure all necessary ports are open and available.
- Verify that Docker, Minikube, and Kubectl are properly configured.
- Troubleshoot any errors by checking logs and verifying configurations.

Congratulations! You have successfully launched the Gobnb accommodation booking platform. Happy hosting and booking!

## Components
- **Client App:** Provides a user interface.
- **Server App:** Microservices, including Auth, Profile, Accommodations, Reservations, Recommendations, Notifications.

## Roles
1. **Unauthenticated User (NK)**
   - Can create a new account or sign in.
   - Can search for accommodations.

2. **Host (H)**
   - Creates and manages accommodations.

3. **Guest (G)**
   - Reserves accommodations.
   - Can rate accommodations and hosts.
     
## Functionalities
- Registration, login, and account management.
- Accommodation creation with details and images.
- Defining availability and prices for accommodations.
- Search for accommodations based on location, guests, and dates.
- Reservation creation and cancellation.
- Rating hosts and accommodations.
- Featured Host status.
- Notifications for hosts.
- Accommodation recommendations for guests.
- Accommodation statistics for hosts.

## System Requirements
1. **Design:** Specified storage, data model, and communication between services.
2. **API Gateway:** Entry point using REST API.
3. **Containerization:** Docker containers using Docker Compose.
4. **Resilience:** System functions if a service is temporarily down.
5. **Tracing:** Implemented tracing with Jeager.
6. **Caching:** Cacheed accommodation images in Redis.
7. **Saga:** Implemented accommodation creation using the saga pattern.
8. **Event Sourcing and CQRS:** Gathered and displayed statistics using these patterns.
9. **Kubernetes:** Runned all components in a Kubernetes cluster.

## Security and Data Protection
1. **Data Validation:** Prevented injection and XSS attacks. Validate data.
2. **HTTPS Communication:** Ensured secure communication.
3. **Authentication and Access Control:** Implemented account verification, RBAC, and access controls.
4. **Data Protection:** Secureed sensitive data during storage, transport, and usage.

## Logging and Vulnerabilities
1. **Completeness:** Logged non-repudiable events and security-related events.
2. **Reliability:** Ensured reliable logging.
3. **Conciseness:** Optimized log entries.
4. **Vulnerabilities:** Identifyed and resolved vulnerabilities. Created a comprehensive report.

## Design of the system
- **Profile Service:** MongoDB - chosen for horizontal scaling, replication support, and dynamic schema flexibility.
- **Notification Service:** MongoDB - dynamic schema and fast data read/write capabilities.
- **Accommodation Service:**
  - **Availability and Prices:** MongoDB - flexible schema for dynamic updates.
  - **Image Storage:** Hadoop Distributed File System (HDFS) - efficient handling of large data volumes.
  - **Image Caching:** Redis - acts as a cache service for quick image data caching in memory.
- **Auth Service:** MongoDB - utilized for authentication service.
- **Recommendation Service:** Neo4j - a graph database for modeling connections between users, accommodations, and interactions.
- **Reservation Service:** Cassandra - chosen for fast distributed processing of time-sensitive data and efficient key-based searches.
- **Rating Service:** MongoDB - universal schema for different types of ratings, ensuring fast read and write performance.
- **Event Sourcing:** Cassandra - event store database with fast write and read capabilities for collecting and displaying statistics.
  
![Gobnb-diagram](https://github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024/assets/96575598/a7c90281-4dd2-4693-acec-5bfd4e6fbd82)

## Images of project

### Registration 
![registration](https://github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024/assets/96575598/4f56b9d9-a50f-4a58-a2fe-54343876942b)
![email verification](https://github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024/assets/96575598/a836f82e-56d3-42bf-88f9-2fe90df2b0de)

### Login 
![login](https://github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024/assets/96575598/a8ab8cd9-a2d8-4ec1-9d04-25078eafca42)

### Forgot password
![forgot password](https://github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024/assets/96575598/8f59955e-a654-4b68-9b0f-03a288d1758a)

### Reset password
![reset password](https://github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024/assets/96575598/4c09886a-ab27-4480-95b0-666238789eb6)

### Main page
![main page host](https://github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024/assets/96575598/957e152e-5fce-45b7-9063-86b050302339)

### Recommendation
![main page and recommendation](https://github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024/assets/96575598/fab0908a-b13f-4025-8eb0-7176fced4d5b)

### Search and filter
![search and filter](https://github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024/assets/96575598/35cff8d2-4f48-418e-a29d-239e7c31ee71)

### Create accommodation
![create accommodation](https://github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024/assets/96575598/29d27ed8-aa05-4895-9136-80c88ef5a983)
![create accommodation 2](https://github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024/assets/96575598/c6b36be1-e680-48c0-9fec-0cddc186711d)

### Create availability for accommodation
![accommodation create availability](https://github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024/assets/96575598/8ddb5ec3-3243-46c6-92fa-223fa14592d0)

### Reserve accommodation and rate host and accommodation
![accommodation reserve](https://github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024/assets/96575598/21d1b8da-164b-46c0-a5d8-5cc434410776)
![accommodation info](https://github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024/assets/96575598/2f709d24-c44d-4bee-9870-e8526537341c)
![accommodation rating](https://github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024/assets/96575598/8042144c-4c42-4acc-ac23-e8080001abff)

### Reservations
![reservations](https://github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024/assets/96575598/0e407d4a-f9d9-4981-a285-d873a5be20b0)

### Ratings
![ratings](https://github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024/assets/96575598/617a2858-d914-4535-8bfa-e803dbb5d2b2)

### Reports
![reports](https://github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024/assets/96575598/28474fa4-19a0-4ffb-8085-fb3fc11ff127)

### Profile
![profile](https://github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024/assets/96575598/6ea99eab-fb4f-486f-8d15-f95ae25780a1)

### Edit profile
![edit profile](https://github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024/assets/96575598/cb5aea44-17eb-4471-afab-76b4307805d0)

### Notifications
![profile notifications](https://github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024/assets/96575598/8489c502-aebe-443e-b666-72ce8730fd0f)






