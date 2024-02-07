# Accommodation Booking Platform - Gobnb

This repository contains the implementation of a platform for offering and booking accommodations. This project is part of the "Service-Oriented Architectures and NoSQL Databases" course.

## Roles
1. **Unauthenticated User (NK)**
   - Can create a new account or sign in.
   - Can search for accommodations.

2. **Host (H)**
   - Creates and manages accommodations.

3. **Guest (G)**
   - Reserves accommodations.
   - Can rate accommodations and hosts.

## Components
- **Client App:** Provides a user interface.
- **Server App:** Microservices, including Auth, Profile, Accommodations, Reservations, Recommendations, Notifications.

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
1. **Design:** Specify storage, data model, and communication between services.
2. **API Gateway:** Entry point using REST API.
3. **Containerization:** Docker containers using Docker Compose.
4. **Resilience:** System functions if a service is temporarily down.
5. **Tracing:** Implement tracing with Jeager.
6. **Caching:** Cache accommodation images in Redis.
7. **Saga:** Implement accommodation creation using the saga pattern.
8. **Event Sourcing and CQRS:** Gather and display statistics using these patterns.
9. **Kubernetes:** Run all components in a Kubernetes cluster.

## Security and Data Protection
1. **Data Validation:** Prevent injection and XSS attacks. Validate data.
2. **HTTPS Communication:** Ensure secure communication.
3. **Authentication and Access Control:** Implement account verification, RBAC, and access controls.
4. **Data Protection:** Secure sensitive data during storage, transport, and usage.

## Logging and Vulnerabilities
1. **Completeness:** Log non-repudiable events and security-related events.
2. **Reliability:** Ensure reliable logging.
3. **Conciseness:** Optimize log entries.
4. **Vulnerabilities:** Identify and resolve vulnerabilities. Create a comprehensive report.

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






