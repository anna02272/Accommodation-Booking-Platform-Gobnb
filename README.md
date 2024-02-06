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
  
![Gobnb-diagram](https://github.com/anna02272/SOA_NoSQL_IB-MRS-2023-2024/assets/96575598/5a5d0c47-9491-4a4a-8bce-a5175052cee8)

