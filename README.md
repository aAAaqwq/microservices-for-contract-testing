# Microservices for Contract Testing

This project demonstrates contract testing in a microservices architecture using Keploy. It includes four core services: user service, order service, payment service, and notification service.

## Architecture Overview
```mermaid
graph TD
    Client --> |HTTP| UserService[User Service<br/>Port:8081<br/>PostgreSQL]
    Client --> |HTTP| OrderService[Order Service<br/>Port:8082<br/>MySQL]
    Client --> |HTTP| PaymentService[Payment Service<br/>Port:8083<br/>PostgreSQL]
    Client --> |HTTP| NotificationService[Notification Service<br/>Port:8084<br/>MongoDB+Redis]
    
    OrderService --> |Validate User| UserService
    OrderService --> |Create Payment| PaymentService
    OrderService --> |Send Notification| NotificationService
    
    PaymentService --> |Validate User| UserService
    PaymentService --> |Validate Order| OrderService
    PaymentService --> |Send Notification| NotificationService
    
    NotificationService --> |Get User Info| UserService
    
    subgraph Databases
        PostgreSQL1[(PostgreSQL<br/>Users)]
        MySQL[(MySQL<br/>Orders)]
        PostgreSQL2[(PostgreSQL<br/>Payments)]
        MongoDB[(MongoDB<br/>Notifications)]
        Redis[(Redis<br/>Message Queue)]
    end
    
    UserService --> PostgreSQL1
    OrderService --> MySQL
    PaymentService --> PostgreSQL2
    NotificationService --> MongoDB
    NotificationService --> Redis
```
## Timing diagram
![alt text](./img/timeline.png)

## API Details
### 1. User Service
- Port: 8081
- Database: PostgreSQL
- Features and APIs:

  - Create User
    ```
    POST /api/v1/users
    Content-Type: application/json
    {
        "username": "string",
        "email": "string",
        "password": "string"
    }
    ```

  - Get User Information
    ```
    GET /api/v1/users/:id
    ```

  - Update User Information
    ```
    PUT /api/v1/users/:id
    Content-Type: application/json
    {
        "username": "string",
        "email": "string"
    }
    ```

  - Delete User
    ```
    DELETE /api/v1/users/:id
    ```

#### Interaction Pattern:
This is a fundamental service providing user registration, query, update, and deletion functionalities for other microservices to interact with.

### 2. Order Service
- Port: 8082
- Database: MySQL
- Features and APIs:

  - Create Order
    ```
    POST /api/v1/orders
    Content-Type: application/json
    {
        "user_id": "uint",
        "items": [
            {
                "name": "string",
                "price": "float64",
                "quantity": "int"
            }
        ]
    }
    ```

  - Get Order Information
    ```
    GET /api/v1/orders/:id
    ```

  - Get User's Orders
    ```
    GET /api/v1/orders/user/:userId
    ```

  - Update Order Status
    ```
    PUT /api/v1/orders/:id/status
    Content-Type: application/json
    {
        "status": "string" // processing, completed, failed, cancelled
    }
    ```

  - Cancel Order
    ```
    DELETE /api/v1/orders/:id
    ```

#### Interaction Pattern:
The order service handles order creation and status management. It interacts with the user service to validate user information, the payment service to process order payments, and the notification service to send order status updates.

### 3. Payment Service
- Port: 8083
- Database: PostgreSQL
- Features and APIs:

  - Create Payment
    ```
    POST /api/v1/payments
    Content-Type: application/json
    {
        "order_id": "uint",
        "user_id": "uint",
        "amount": "float64",
        "payment_type": "string" // credit_card, alipay, wechat
    }
    ```

  - Get Payment Information
    ```
    GET /api/v1/payments/:id
    ```

  - Get Order Payment Information
    ```
    GET /api/v1/payments/order/:orderId
    ```

  - Get User Payment Records
    ```
    GET /api/v1/payments/user/:userId
    ```

  - Process Refund
    ```
    POST /api/v1/payments/:id/refund
    Content-Type: application/json
    {
        "reason": "string"
    }
    ```

#### Interaction Pattern:
The payment service handles order payment and refund functionalities. It interacts with the order service to validate order information, the user service to verify user information, and the notification service to send payment status notifications.

### 4. Notification Service
- Port: 8084
- Database: MongoDB + Redis
- Features and APIs:

  - Create Notification
    ```
    POST /api/v1/notifications
    Content-Type: application/json
    {
        "user_id": "uint",
        "type": "string",     // email, sms
        "title": "string",
        "content": "string",
        "recipient": "string" // email address or phone number
    }
    ```

  - Get Notification Information
    ```
    GET /api/v1/notifications/:id
    ```

  - Get User's Notifications
    ```
    GET /api/v1/notifications/user/:userId
    ```

  - Batch Create Notifications
    ```
    POST /api/v1/notifications/batch
    Content-Type: application/json
    {
        "notifications": [
            {
                "user_id": "uint",
                "type": "string",
                "title": "string",
                "content": "string",
                "recipient": "string"
            }
        ]
    }
    ```

#### Interaction Pattern:
The notification service handles all notification sending functionalities in the system. It receives notification requests from other services (order, payment, etc.), uses Redis as a message queue for asynchronous notification processing, and stores notification records in MongoDB. It supports both email and SMS notification methods.

## Getting Started



### Preparation
1. Clone the repository and enter the directory
2. Install keploy-cli: 
  ```
  curl --silent -O -L https://keploy.io/install.sh && source install.sh
  ```

### Run app using docker
Keploy can be used on Linux, Windows and MacOS through Docker.
> Note: To run Keploy on MacOS through Docker the version must be 4.25.2 or above.
1. Run `docker compose up` to start all db-instances
2. Create the docker image for each service
  ```
  docker build -t user-service:1.0 ./user-service
  ```
  ```
  docker build -t order-service:1.0 ./order-service
  ```
  ```
  docker build -t payment-service:1.0 ./payment-service
  ```
  ```
  docker build -t notification-service:1.0 ./notification-service
  ```
3. Run the docker images
  ```
  keploy record -c "docker run -p 8081:8081 userApp -n microservices-network user-service:1.0"
  ```
  ```
  keploy record -c "docker run -p 8082:8082 orderApp -n microservices-network order-service:1.0"
  ```
  ```
  keploy record -c "docker run -p 8083:8083 paymentApp -n microservices-network payment-service:1.0"
  ```
  ```
  keploy record -c "docker run -p 8084:8084 notificationApp -n microservices-network notification-service:1.0"
  ```
  >Tips:Keploy can only record one service at a time. When you record a service, other services must be running.
  
  To genereate testcases we just need to make some API calls. You can use Postman, Hoppscotch, or simply curl



### Keploy Contract Testing Results Analysis
<!-- #### Testing
Each service includes contract tests using Keploy. To run the tests:
1. Navigate to the service directory
2. Run `keploy test -c "go run ." -d 10`
3. If you want to see test coverage, run `keploy test -c ./<service-name> -d 10` -->
   
#### My testing data
- user-service
![alt text](./img/user-srv.png)

- order-service
![alt text](./img/order-srv.png)
![alt text](./img/order-fail2.png)
![alt text](./img/order-fail5.png)
![alt text](./img/order-fail7.png)
-payment-service
![alt text](./img/payment-srv.png)
![alt text](./img/payment-fail2.png)
-notification-service
![alt text](./img/notification-srv.png)
![alt text](./img/notification-fail1.png)
![alt text](./img/notification-fail2.png)
![alt text](./img/notification-fail3.png)

#### Test efficiency analysis
1. Capture dependencies between services
2. Asynchronous operation processing
3. Data storage interaction 

### Potential limitations
1. It is difficult to record multiple services at the same time
2. Time-related operations are difficult to simulate and reproduce
3. The contract generated by keploy simulation cannot be managed persistently
4. Keploy cannot simulate external service calls such as smtp.
5. Dynamic ID generation → Response body length changes → Content-Length does not match → Test assertion fails
6. Timestamp time zone difference → String length fluctuation → Response body byte count changes → Header verification fails

## Possible optimization points for microservices
- Use gRPC for communication between microservices to improve throughput
- Use RocketMQ for message queues to improve message processing efficiency
- Use sonic tools for json processing to improve performance
- Use snowflake algorithm for ID generation to improve uniqueness
- ....


### Technology stack
- Language: Go 1.23.3
- Web framework: Gin
- Database:
  - PostgreSQL
  - MySQL
  - Redis
  - Mongo
- Containerization: Docker
- API style: RESTful
- Test framework: Keploy
- Deployment: Docker Compose

## License
This project is licensed under the MIT License - see the LICENSE file for details
